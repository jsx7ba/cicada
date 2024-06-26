package cicada

// Room indicates who can communicate with each other.
type Room struct {
	Id          string   `clover:"id" json:"id,omitempty"`
	Name        string   `clover:"name" json:"name"`
	Description string   `clover:"description" json:"description"`
	Members     []string `clover:"members" json:"members,omitempty"`
}
