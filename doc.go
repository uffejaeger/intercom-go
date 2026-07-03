// Package intercom provides an idiomatic Go client for the Intercom API.
//
// Create a client with an Intercom access token:
//
//	client, err := intercom.NewClient("access-token")
//
// Or read the token from INTERCOM_ACCESS_TOKEN:
//
//	client, err := intercom.NewClientFromEnv()
//
// The client exposes service fields for Intercom resources, for example:
//
//	admin, err := client.Admins.Me(ctx)
//	contact, err := client.Contacts.Get(ctx, "contact_id")
//	conversations, err := client.Conversations.List(ctx)
package intercom
