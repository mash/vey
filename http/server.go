package http

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/mash/vey"
	"github.com/mash/vey/email"
)

type VeyHandler struct {
	*http.ServeMux
	Vey     vey.Vey
	Sender  email.Sender
	OpenURL *url.URL
}

func NewHandler(vey vey.Vey, sender email.Sender, open *url.URL) http.Handler {
	h := VeyHandler{
		ServeMux: http.NewServeMux(),
		Vey:      vey,
		Sender:   sender,
		OpenURL:  open,
	}
	h.Handle("/getKeys", WrapF(AcceptJSON(h.GetKeys)))
	h.Handle("/beginDelete", WrapF(AcceptJSON(h.BeginDelete)))
	h.Handle("/commitDelete", WrapF(h.CommitDelete))
	h.Handle("/beginPut", WrapF(AcceptJSON(h.BeginPut)))
	h.Handle("/commitPut", WrapF(AcceptJSON(h.CommitPut)))
	h.Handle("/open", WrapF(h.Open))
	return &h
}

type Body struct {
	Email     string        `json:"email,omitempty"`
	PublicKey vey.PublicKey `json:"publickey,omitempty"`
	Token     []byte        `json:"token,omitempty"`
	Challenge []byte        `json:"challenge,omitempty"`
	Signature []byte        `json:"signature,omitempty"`
}

func (h *VeyHandler) GetKeys(w http.ResponseWriter, r *http.Request, b Body) error {
	keys, err := h.Vey.GetKeys(b.Email)
	if err != nil {
		return err
	}
	return WriteJSON(w, 200, keys)
}

func (h *VeyHandler) BeginDelete(w http.ResponseWriter, r *http.Request, b Body) error {
	token, err := h.Vey.BeginDelete(b.Email, b.PublicKey)
	if err != nil {
		return err
	}
	if err := h.Sender.SendToken(b.Email, base64.StdEncoding.EncodeToString(token)); err != nil {
		return err
	}
	return WriteJSON(w, 200, map[string]interface{}{})
}

// CommitDelete handles the final step of deleting the public key.
// The user receives the URL to CommitDelete in the email and opens it in their browser.
// token parameter should be in query.
func (h *VeyHandler) CommitDelete(w http.ResponseWriter, r *http.Request) error {
	s := r.URL.Query().Get("token")
	token, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Error{
			Code: 400,
			Msg:  "token decode failed",
			Err:  err,
		}
	}
	if err := h.Vey.CommitDelete(token); err != nil {
		return err
	}
	return WriteJSON(w, 200, map[string]interface{}{})
}

func (h *VeyHandler) BeginPut(w http.ResponseWriter, r *http.Request, b Body) error {
	challenge, err := h.Vey.BeginPut(b.Email)
	if err != nil {
		return err
	}
	if err := h.Sender.SendChallenge(b.Email, base64.StdEncoding.EncodeToString(challenge)); err != nil {
		return err
	}
	return WriteJSON(w, 200, map[string]interface{}{})
}

func (h *VeyHandler) CommitPut(w http.ResponseWriter, r *http.Request, b Body) error {
	if err := h.Vey.CommitPut(b.Challenge, b.Signature, b.PublicKey); err != nil {
		if err == vey.ErrVerifyFailed {
			return Error{
				Code: 400,
				Msg:  err.Error(),
				Err:  nil,
			}
		} else if err == vey.ErrNotFound {
			return Error{
				Code: 404,
				Msg:  err.Error(),
				Err:  nil,
			}
		}
		return err
	}
	return WriteJSON(w, 200, map[string]interface{}{})
}

func (h *VeyHandler) Open(w http.ResponseWriter, r *http.Request) error {
	if h.OpenURL == nil || r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}

	next := *h.OpenURL
	next.RawQuery = r.URL.RawQuery
	http.Redirect(w, r, next.String(), http.StatusFound)
	return nil
}
