package email

import (
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
)

// See https://pkg.go.dev/github.com/aws/aws-sdk-go/service/ses#SendTemplatedEmailInput for details.
type SESConfig struct {
	// The email address that is sending the email. This email address must be either
	// individually verified with Amazon SES, or from a domain that has been verified
	// with Amazon SES. For information about verifying identities, see the Amazon
	// SES Developer Guide (https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html).
	Source string `yaml:"source"`
	// This parameter is used only for sending authorization. It is the ARN of the
	// identity that is associated with the sending authorization policy that permits
	// you to send for the email address specified in the Source parameter.
	SourceArn string `yaml:"sourceArn"`
	// The name of the configuration set to use when you send an email using SendTemplatedEmail.
	// If you do not want to use a configuration set, set to empty string.
	ConfigurationSetName string `yaml:"configurationSetName"`
	// The reply-to email address(es) for the message. If the recipient replies
	// to the message, each reply-to address will receive the reply.
	ReplyToAddresses []string `yaml:"replyToAddresses"`
	// The email address that bounces and complaints will be forwarded to when feedback
	// forwarding is enabled. If the message cannot be delivered to the recipient,
	// then an error message will be returned from the recipient's ISP; this message
	// will then be forwarded to the email address specified by the ReturnPath parameter.
	// The ReturnPath parameter is never overwritten. This email address must be
	// either individually verified with Amazon SES, or from a domain that has been
	// verified with Amazon SES.
	ReturnPath string `yaml:"returnPath"`
	// This parameter is used only for sending authorization. It is the ARN of the
	// identity that is associated with the sending authorization policy that permits
	// you to use the email address specified in the ReturnPath parameter.
	ReturnPathArn string `yaml:"returnPathArn"`
	// AWS SES email template to use when sending the delete confirmation email.
	// The {{email}}, {{token}} and {{tokenEscaped}} variables in the template are replaced with values set by Vey.
	DeleteTemplate string `yaml:"deleteTemplate"`
	// AWS SES email template to use when sending the put confirmation email.
	// The {{email}}, {{challenge}} and {{challengeEscaped}} variables in the template are replaced with values set by Vey.
	PutTemplate string `yaml:"putTemplate"`
}

// NewSESSender returns a SESSender which sends email via Amazon SES.
// If the templates have an error, NewSESSender may panic.
func NewSESSender(c SESConfig, ses *ses.SES) Sender {
	return SESSender{
		Config: c,
		SES:    ses,
	}
}

// SESSender implements Sender interface using AWS SES.
type SESSender struct {
	Config SESConfig
	SES    *ses.SES
}

type sesData struct {
	Email            string `json:"email"`
	Token            string `json:"token,omitempty"`
	TokenEscaped     string `json:"tokenEscaped,omitempty"`
	Challenge        string `json:"challenge,omitempty"`
	ChallengeEscaped string `json:"challengeEscaped,omitempty"`
}

// SendToken sends the token to the dst email address.
func (s SESSender) SendToken(dst, token string) error {
	// use tokenEscaped in template if token is added in query parameter in the template.
	data := sesData{
		Email:        dst,
		Token:        token,
		TokenEscaped: url.QueryEscape(token),
	}
	return s.send(dst, "delete", "vey_delete", data)
}

// SendChallenge sends the challenge to the dst email address.
func (s SESSender) SendChallenge(dst, challenge string) error {
	// use challengeEscaped in template if token is added in query parameter.
	data := sesData{
		Email:            dst,
		Challenge:        challenge,
		ChallengeEscaped: url.QueryEscape(challenge),
	}
	return s.send(dst, "put", "vey_put", data)
}

func (s SESSender) send(email, action, template string, data sesData) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	input := &ses.SendTemplatedEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Source:       aws.String(s.Config.Source),
		Template:     aws.String(template),
		TemplateData: aws.String(string(j)),
		Tags: []*ses.MessageTag{
			{
				Name:  aws.String("action"),
				Value: aws.String(action),
			},
		},
	}
	if s.Config.SourceArn != "" {
		input.SourceArn = aws.String(s.Config.SourceArn)
	}
	if s.Config.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.Config.ConfigurationSetName)
	}
	if s.Config.ReplyToAddresses != nil && len(s.Config.ReplyToAddresses) > 0 {
		input.ReplyToAddresses = aws.StringSlice(s.Config.ReplyToAddresses)
	}
	if s.Config.ReturnPath != "" {
		input.ReturnPath = aws.String(s.Config.ReturnPath)
	}
	if s.Config.ReturnPathArn != "" {
		input.ReturnPathArn = aws.String(s.Config.ReturnPathArn)
	}

	// Attempt to send the email.
	_, err = s.SES.SendTemplatedEmail(input)
	return err
}
