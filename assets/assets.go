package assets

import "embed"

//go:embed templates/* locales/*
var Assets embed.FS
