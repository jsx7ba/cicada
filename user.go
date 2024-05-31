package cicada

// User defines information for a chat user.
type User struct {
	Id    string   `clover:"id" json:"id,omitempty"`
	Name  string   `clover:"name" json:"name"`
	Rooms []string `clover:"rooms" json:"rooms"`
}
