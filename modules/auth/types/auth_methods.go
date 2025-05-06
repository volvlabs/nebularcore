package types

// AuthMethod represents an authentication method
type AuthMethod string

const (
	// MethodUsernamePassword represents username/password authentication
	MethodUsernamePassword AuthMethod = "username_password"
	// MethodEmailPassword represents email/password authentication
	MethodEmailPassword AuthMethod = "email_password"
	// MethodPhonePassword represents phone/password authentication
	MethodPhonePassword AuthMethod = "phone_password"
)
