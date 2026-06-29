package cache

// Subject represents a resolved subject (User, IdpIdentity, IdpClient).
// The entity being authenticated.
type Subject struct {
	Key        string
	Email      string
	Scopes     []string
	UserId     *uint
	IdentityId *uint
	ClientId   *uint
	User       *User
	Identity   *Identity
	Client     *IdpClient
	Task       *Task
}

// WithUser populates Subject from a User model.
func (r *Subject) WithUser(user *User, scopes []string) {
	r.UserId = &user.ID
	r.Key = user.Subject
	r.User = user
	r.Email = user.Email
	r.Scopes = scopes
}

// WithIdentity populates Subject from an IdpIdentity model.
func (r *Subject) WithIdentity(idp *Identity) {
	r.IdentityId = &idp.ID
	r.Identity = idp
	r.Key = idp.Subject
	r.Email = idp.Email
	r.Scopes = idp.Scopes
}

// WithClient populates Subject from an IdpClient model.
func (r *Subject) WithClient(client *IdpClient) {
	r.ClientId = &client.ID
	r.Client = client
	r.Key = client.Subject
	r.Scopes = client.GetScopes()
}

// WithTask populates with the task.
func (r *Subject) WithTask(task *Task) {
	r.Task = task
	r.Key = task.Subject()
	r.Scopes = task.GetScopes()
}

// Login returns the user (login) name (Eg: jsmith).
func (r *Subject) Login() (login string) {
	if r.IsUser() {
		login = r.User.Login
		return
	}
	if r.IsIdentity() {
		login = r.Identity.Login
		return
	}
	if r.IsClient() {
		login = r.Client.ClientId
		return
	}
	if r.IsTask() {
		login = r.Task.Login()
	}
	return
}

// IsUser returns true if this subject is a User.
func (r *Subject) IsUser() bool {
	return r.UserId != nil
}

// IsIdentity returns true if this subject is an IdpIdentity.
func (r *Subject) IsIdentity() bool {
	return r.IdentityId != nil
}

// IsClient returns true if this subject is an IdpClient.
func (r *Subject) IsClient() bool {
	return r.ClientId != nil
}

// IsTask returns true if the subject is a Task.
func (r *Subject) IsTask() bool {
	return r.Task != nil
}
