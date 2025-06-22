package models

type State string

type StateGroup string

func (s State) Group() StateGroup {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return StateGroup(s[:i])
		}
	}
	return ""
}


const (
	DefaultState       State = ""
	StateStart         State = "start"
	StateAwaitingLogin State = "awaiting_login"
	StateAwaitingOTP   State = "awaiting_otp"
	StateRegistered    State = "registered"
)