module github.com/mash/vey/cmd/vey

go 1.17

require (
	github.com/joho/godotenv v1.4.0
	github.com/rs/zerolog v1.26.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	github.com/mash/vey v0.0.0
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
)

replace (
	github.com/mash/vey v0.0.0 => ../..
)
