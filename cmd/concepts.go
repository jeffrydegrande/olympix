package cmd

// DefaultSecurityConcepts returns the default security concepts without embeddings
func DefaultSecurityConcepts() []SecurityConcept {
	return []SecurityConcept{
		{
			Name:        "active",
			Description: "Concept representing whether a market is active/initialized",
			Synonyms:    []string{"enabled", "live", "activated", "initialized", "ready"},
		},
		{
			Name:        "locked",
			Description: "Concept representing a reentrancy guard or mutex",
			Synonyms:    []string{"reentrancy_guard", "mutex", "guard", "semaphore"},
		},
		{
			Name:        "grace_period",
			Description: "Concept representing a waiting period or timelock",
			Synonyms:    []string{"timelock", "delay", "cooldown", "waiting_period"},
		},
		{
			Name:        "admin",
			Description: "Concept representing an administrative role or owner",
			Synonyms:    []string{"owner", "administrator", "authority", "governor"},
		},
		{
			Name:        "min_deposit",
			Description: "Concept representing a minimum deposit or liquidity threshold",
			Synonyms:    []string{"minimum_deposit", "min_liquidity", "threshold", "minimum_amount"},
		},
		{
			Name:        "bounds_check",
			Description: "Concept representing bounds checking or validations",
			Synonyms:    []string{"validation", "assert", "check", "limit", "cap"},
		},
		{
			Name:        "accumulator",
			Description: "Concept representing an accumulator or counter",
			Synonyms:    []string{"counter", "index", "tally", "tracker"},
		},
		{
			Name:        "donation_cap",
			Description: "Concept representing a cap on donations or contributions",
			Synonyms:    []string{"cap", "limit", "maximum", "ceiling"},
		},
	}
}

