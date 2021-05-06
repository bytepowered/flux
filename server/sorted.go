package server

import (
	"github.com/bytepowered/flux"
	"sort"
)

func sortedStartup(items []flux.Startuper) []flux.Startuper {
	out := make(StartupArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}

func sortedShutdown(items []flux.Shutdowner) []flux.Shutdowner {
	out := make(ShutdownArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}

type StartupArray []flux.Startuper

func (s StartupArray) Len() int           { return len(s) }
func (s StartupArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StartupArray) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

type ShutdownArray []flux.Shutdowner

func (s ShutdownArray) Len() int           { return len(s) }
func (s ShutdownArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ShutdownArray) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

func orderOf(v interface{}) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
