package server

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"sort"
)

func sortedStartup(items []flux.Startuper) []flux.Startuper {
	out := make(StartupByOrderer, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Stable(out)
	return out
}

func sortedShutdown(items []flux.Shutdowner) []flux.Shutdowner {
	out := make(ShutdownByOrderer, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Stable(out)
	return out
}

type StartupByOrderer []flux.Startuper

func (s StartupByOrderer) Len() int           { return len(s) }
func (s StartupByOrderer) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StartupByOrderer) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

type ShutdownByOrderer []flux.Shutdowner

func (s ShutdownByOrderer) Len() int           { return len(s) }
func (s ShutdownByOrderer) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ShutdownByOrderer) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

func orderOf(v interface{}) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
