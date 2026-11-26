// Package cards holds the card definitions built from ability shapes
// (ADR-0004) and the allowlist of cards that cannot run yet.
// It imports only package ability.
package cards

import "animus-mundi/pkg/engine/ability"

// Defs maps a card name to its ability definition.
var Defs = map[string]*ability.Def{}

// NotYetExecutable lists, sorted, every card in data/cards the kernel cannot
// execute yet (pkg/cardcheck's TestCards_Executable). Card tasks remove names
// from it; M3 is done when it is empty.
var NotYetExecutable = []string{
	"All-Out War",
	"Armored Charge",
	"Awaken Nightmare",
	"Cathedral of Self",
	"Command Barge",
	"Comms Relay",
	"Consuming Earth",
	"Dolus Infusion",
	"Dolus Macrocannon",
	"Dolus Refiner",
	"Dominator Warsuit",
	"Dread Manifestation",
	"Earthshaker Colossus",
	"Encampment",
	"Enclave",
	"Field Mechanic",
	"Fire for Effect",
	"Flickering Ember",
	"Forgemaster's Resilience",
	"Frostfire Breath",
	"Gemstone Titan",
	"Hellfire Armorers",
	"Hourglass Diviner",
	"Inspired Salvage",
	"Land Fortress",
	"Light of the World Soul",
	"MAX-8 Autonomous Warsuit",
	"Macrocannon Barrage",
	"Mage Seeker's Inquisition",
	"March of Progress",
	"Mobile Armory",
	"NO-AH-5",
	"Natural Attunement",
	"Piston Fist",
	"Piston Fist Prototype",
	"Popup Bunker",
	"Rage's Wake",
	"Railway Reinforcements",
	"Redline Infusion",
	"Redlined Brawler",
	"Reverberating Terror",
	"Rift",
	"Samsara, Living Tempest",
	"Sandstorm Lightbringer",
	"Screaming Palisade",
	"Stardust Colossus",
	"Tempest Runner",
	"Tempest Surge",
	"Tidal Leviathan",
	"Visionary's Eye",
	"Visionary's Insight",
	"World Soul Messenger",
}
