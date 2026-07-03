// Package intercomtest provides offline test helpers for code using intercom-go.
//
// The package is built around httptest and does not call Intercom. Tests script
// expected routes, run SDK calls against the local server, and then inspect the
// captured requests.
//
// Example:
//
//	srv := intercomtest.NewServer(t,
//		intercomtest.Route("GET", "/me", intercomtest.JSON(200, `{
//			"type": "admin",
//			"id": "admin-1",
//			"email": "admin@example.com"
//		}`)),
//	)
//
//	client, err := srv.Client("token")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	admin, err := client.Admins.Me(context.Background())
//	if err != nil {
//		t.Fatal(err)
//	}
//	if admin.Email == nil || *admin.Email != "admin@example.com" {
//		t.Fatalf("unexpected admin email: %v", admin.Email)
//	}
//
//	req := srv.Request(t, 0)
//	if req.Method != "GET" || req.Path != "/me" {
//		t.Fatalf("unexpected request: %s %s", req.Method, req.Path)
//	}
package intercomtest
