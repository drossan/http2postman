package model

// Environment represents the contents of http-client.env.json.
// Outer key is the environment name, inner map contains the variables.
type Environment map[string]map[string]string
