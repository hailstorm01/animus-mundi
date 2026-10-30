//go:build !windows

package main

// keepAwake does nothing outside Windows.
func keepAwake() {}
