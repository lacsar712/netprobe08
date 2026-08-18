package policy

import x "example.com/netprobe/internal/twamp"

type Item struct {
	Title, Body string
	Tags        []string
}

func Enforce(title, body string, tags []string) error {
	return x.Enforce(title, body, tags)
}

func Sample() Item {
	s := x.Sample()
	return Item{Title: s.Title, Body: s.Body, Tags: s.Tags}
}

func Seed() []Item {
	var out []Item
	for _, s := range x.Seed() {
		out = append(out, Item{Title: s.Title, Body: s.Body, Tags: s.Tags})
	}
	return out
}

func AfterWrite(getMin func() (string, error), setMin func(string) error, body string) error {
	return x.AfterWrite(getMin, setMin, body)
}

func Steps() []string { return x.Steps() }
