module kuassh

go 1.14

require (
	github.com/manifoldco/promptui v0.7.0
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/nsf/termbox-go v0.0.0-20200418040025-38ba6e5628f1
	github.com/spf13/cobra v1.0.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/spf13/cobra v1.0.0 => github.com/spf13/cobra v1.0.0

replace github.com/manifoldco/promptui v1.0.0 => ./vendor/github.com/manifoldco/promptui
