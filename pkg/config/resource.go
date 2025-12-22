package config

import "sync"

type Resource struct {
	Whitelist                 []string `yaml:"whitelist"`
	Blacklist                 []string `yaml:"blacklist"`
	SkipEvents                []string `yaml:"skipEvents"`
	IgnoreFetchErrorResources []string `yaml:"ignoreFetchErrorResources"`

	whitelist []string
	once      sync.Once

	mpIgnoreFetchErrorResources  map[string]struct{}
	ignoreFetchErrorResourceOnce sync.Once
}

func (r *Resource) GetWhitelist() []string {
	r.once.Do(func() {
		mpBlacklist := map[string]struct{}{}
		for _, value := range r.Blacklist {
			mpBlacklist[value] = struct{}{}
		}
		for _, value := range r.Whitelist {
			// Blacklistに存在するキーは対象外とする
			if _, ok := mpBlacklist[value]; !ok {
				r.whitelist = append(r.whitelist, value)
			}
		}
	})
	return r.whitelist
}

func (r *Resource) GetBlacklist() []string {
	return r.Blacklist
}

func (r *Resource) GetSkipEvents() []string {
	return r.SkipEvents
}

func (r *Resource) IsIgnoreFetchErrorResources(resource string) bool {
	if r.mpIgnoreFetchErrorResources == nil {
		r.mpIgnoreFetchErrorResources = map[string]struct{}{}
	}
	r.ignoreFetchErrorResourceOnce.Do(func() {
		for _, rs := range r.IgnoreFetchErrorResources {
			r.mpIgnoreFetchErrorResources[rs] = struct{}{}
		}
	})
	_, ok := r.mpIgnoreFetchErrorResources[resource]
	return ok
}
