package cicada

// Image represents image metadata stored inside a chat log.
type Image struct {
	Id          string `clover:"id" json:"id"`
	Name        string `clover:"name" json:"name"`
	ContentType string `clover:"contentType" json:"contentType"`
}
