// Package customers is the people the artist draws for and sells to, and the social networks she
// reaches them on. The networks are a list she edits: each has a link pattern that turns a
// handle into the address of the profile.
package customers

import "strings"

// UserMark is where a network's link pattern takes the handle: "https://x.com/{user}".
const UserMark = "{user}"

// Network is a social network the artist reaches customers on.
type Network struct {
	ID          string
	Name        string
	LinkPattern string
	Icon        string
}

// Link is the address of a handle's profile on the network; the "@" people write is left out.
func (network Network) Link(handle string) string {
	return strings.ReplaceAll(network.LinkPattern, UserMark, strings.TrimPrefix(strings.TrimSpace(handle), "@"))
}

// Profile is a customer's handle on one network.
type Profile struct {
	NetworkID string
	Handle    string
}

// Customer is someone the artist works for: a name, a way to reach them and their profiles.
type Customer struct {
	ID       string
	Name     string
	Contact  string // an email or a phone, as typed
	Profiles []Profile
}

// Handle is the customer's handle on a network; found is false when they have none there.
func (customer Customer) Handle(networkID string) (handle string, found bool) {
	for _, profile := range customer.Profiles {
		if profile.NetworkID == networkID && profile.Handle != "" {
			return profile.Handle, true
		}
	}
	return "", false
}
