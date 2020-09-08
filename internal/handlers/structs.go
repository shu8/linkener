package handlers

type usernameContext string

// UsernameContextKey - identify the http Context for the username that is passed into handlers
var UsernameContextKey = usernameContext("username")
