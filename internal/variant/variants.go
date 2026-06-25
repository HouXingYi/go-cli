package variant

// Ent is the enterprise management CLI (formerly zn-cli).
var Ent = Variant{
	ID:      "ent",
	AppName: "zn-ent",
	CLIType: "ent",
	StateDir: "zn-ent",
}

// Eco is the ecosystem CLI.
var Eco = Variant{
	ID:      "eco",
	AppName: "zn-eco",
	CLIType: "eco",
	StateDir: "zn-eco",
}
