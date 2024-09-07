package main

import "C"

type ProfileState string

const (
	ProfileStateEnabled  ProfileState = "enabled"
	ProfileStateDisabled ProfileState = "disabled"
)

var CError C.int = -1
