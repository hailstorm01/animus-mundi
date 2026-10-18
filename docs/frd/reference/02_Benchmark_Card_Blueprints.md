# Canonical Benchmark Card Blueprint Implementations
**Companion Reference Catalog for [`02_Card_Data_Pipeline_FRD.md`](../02_Card_Data_Pipeline_FRD.md) (Section 4.2)**

This reference document contains the 14 canonical `CardBlueprint` JSON implementations demonstrating the declarative Ability AST (`FR-ABL-001` – `FR-ABL-012`), event hook bindings (`FR-TRG-001`, `FR-TRG-007`), atomic effect primitives (`FR-PRM-001`), and upfront target locking rules (`FR-TGT-101`).

---

#### 1. `NO-AH-5` (Soulless Mythic Unit - Tucking & Composable Query Scaling)
```json
{
  "id": "soulless_no_ah_5",
  "name": "NO-AH-5",
  "cardType": "UNIT",
  "faction": "SOULLESS",
  "subtypes": ["Officer", "Cybernetic"],
  "requisitionCost": 3,
  "tacticsCost": 0,
  "power": 2,
  "health": 5,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "Ward\nR2: Kill chosen allied non-Headquarters asset and tuck it face down under this unit. This unit has +2 power for each card tucked under it.",
  "abilities": [
    { "abilityType": "KEYWORD_GRANT", "keyword": "WARD" },
    {
      "abilityType": "ACTIVATED",
      "costs": [
        { "costType": "RESOURCE", "resource": "REQUISITION", "amount": 2 },
        { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "ASSET", "excludeHQ": true } }
      ],
      "effect": { "primitive": "TUCK_UNDER", "hostTarget": { "selector": "SELF" }, "faceDown": true }
    },
    {
      "abilityType": "CONTINUOUS_AURA",
      "scope": "SELF",
      "modifiers": [
        {
          "attribute": "COMBAT_POWER",
          "operation": "ADD",
          "value": {
            "type": "QUERY",
            "source": "SELF",
            "query": { "zone": "ATTACHED", "aggregate": "COUNT" },
            "multiplier": 2
          }
        }
      ]
    }
  ]
}
```

#### 2. `Piston Fist Prototype` (Soulless Mythic Unit - 6-Drop Pneumatic Recoil)
```json
{
  "id": "soulless_piston_fist_prototype",
  "name": "Piston Fist Prototype",
  "cardType": "UNIT",
  "faction": "SOULLESS",
  "subtypes": ["Mechanized"],
  "requisitionCost": 3,
  "tacticsCost": 3,
  "power": 1,
  "health": 9,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "Aggressive 9\nAfter this unit attacks, move this unit to your Reserves.",
  "abilities": [
    { "abilityType": "KEYWORD_GRANT", "keyword": "AGGRESSIVE", "parameter": 9 },
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_ATTACK_RESOLVED",
      "condition": { "source": "SELF" },
      "effect": {
        "primitive": "MOVE_ZONE",
        "destinationZone": "RESERVES"
      }
    }
  ]
}
```

#### 3. `Animoceps Spore Cloud` (Animus Rare Structure - Replacement Redirection)
```json
{
  "id": "animus_animoceps_spore_cloud",
  "name": "Animoceps Spore Cloud",
  "cardType": "STRUCTURE",
  "faction": "ANIMUS",
  "subtypes": ["Fungus"],
  "requisitionCost": 4,
  "tacticsCost": 0,
  "power": 0,
  "health": 9,
  "requisitionYield": 2,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER"],
  "rarity": "RARE",
  "rulesText": "Redirect all damage that would be dealt to allied units in the same zone to this structure.",
  "abilities": [
    {
      "abilityType": "REPLACEMENT_WATCHER",
      "interceptHook": "WOULD_TAKE_DAMAGE",
      "condition": { "allegiance": "ALLIED", "entityType": "UNIT", "zone": "SAME_ZONE" },
      "replacementAction": { "primitive": "REDIRECT_DAMAGE_TARGET", "newTarget": { "selector": "SELF" } }
    }
  ]
}
```

#### 4. `Samsara, Living Tempest` (Concordia Mythic Unit - Upfront Bound Trigger & Composite Partition)
```json
{
  "id": "concordia_samsara_living_tempest",
  "name": "Samsara, Living Tempest",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Elemental", "Tempest"],
  "requisitionCost": 6,
  "tacticsCost": 0,
  "power": 6,
  "health": 7,
  "requisitionYield": 0,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER"],
  "rarity": "MYTHIC",
  "rulesText": "Whenever this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard pile. If the card put into your discard pile this way was a maneuver card, deal 2 damage to chosen asset.",
  "abilities": [
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_ATTACK_DECLARE",
      "condition": { "source": "SELF" },
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "entityType": "ASSET" },
        "count": 1,
        "bindingPolicy": "UPFRONT_LOCK"
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "LOOK_TOP_CARDS",
            "count": 2,
            "partition": {
              "destinations": [
                { "zone": "HAND", "count": 1 },
                { "zone": "DISCARD", "count": 1, "trackAs": "discardedCard" }
              ]
            }
          },
          {
            "primitive": "CONDITIONAL_CLAUSE",
            "condition": {
              "targetRef": "discardedCard",
              "cardType": "MANEUVER"
            },
            "effect": {
              "primitive": "DEAL_DAMAGE",
              "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
              "amount": { "type": "STATIC", "value": 2 }
            }
          }
        ]
      }
    }
  ]
}
```

#### 5. `Swirling Dustdevil` (Concordia Common Unit - Hand-Active Self Cost Reduction)
```json
{
  "id": "concordia_swirling_dustdevil",
  "name": "Swirling Dustdevil",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Elemental"],
  "requisitionCost": 8,
  "tacticsCost": 0,
  "power": 6,
  "health": 5,
  "requisitionYield": 2,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "COMMON",
  "rulesText": "This unit costs R1 less to deploy for each maneuver card in your discard pile.",
  "abilities": [
    {
      "abilityType": "CONTINUOUS_AURA",
      "activeZone": "HAND",
      "scope": "SELF",
      "modifiers": [
        {
          "attribute": "REQUISITION_COST",
          "operation": "SUBTRACT",
          "value": {
            "type": "QUERY",
            "source": "GAME_STATE",
            "query": {
              "zone": "DISCARD",
              "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "MANEUVER" },
              "aggregate": "COUNT"
            },
            "multiplier": 1
          }
        }
      ]
    }
  ]
}
```

#### 6. `Veil Render` (Animus Mythic Unit - Discard-Pile Static Deployment Permission)
```json
{
  "id": "animus_veil_render",
  "name": "Veil Render",
  "cardType": "UNIT",
  "faction": "ANIMUS",
  "subtypes": ["Spirit", "Horror"],
  "requisitionCost": 3,
  "tacticsCost": 3,
  "power": 5,
  "health": 4,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "You may deploy this card from your discard pile by killing an allied asset in addition to paying its deployment costs.",
  "abilities": [
    {
      "abilityType": "CONTINUOUS_AURA",
      "activeZone": "DISCARD",
      "scope": "SELF",
      "staticPermission": {
        "permissionType": "ALLOW_DEPLOY_FROM_ZONE",
        "sourceZone": "DISCARD",
        "additionalCosts": [
          { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "ASSET" } }
        ]
      }
    }
  ]
}
```

#### 7. `Rage's Wake` (Concordia Common Maneuver - Delayed Trigger & Forced Free-Deploy Zone)
```json
{
  "id": "concordia_rages_wake",
  "name": "Rage's Wake",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "COMMON",
  "rulesText": "Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "entityType": "UNIT" },
        "count": 1
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "MODIFY_STAT",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "attribute": "COMBAT_POWER",
            "amount": { "type": "STATIC", "value": 2 },
            "duration": "END_OF_TURN"
          },
          {
            "primitive": "REGISTER_DELAYED_TRIGGER",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "triggerHook": "EVENT_DESTROY",
            "duration": "END_OF_TURN",
            "condition": { "source": "SELF" },
            "effect": {
              "primitive": "DEPLOY_FREE",
              "sourceZone": "DISCARD",
              "destinationZone": "RESERVES",
              "target": { "selector": "SELF" }
            }
          }
        ]
      }
    }
  ]
}
```

#### 8. `Funeral Fungus` (Animus Uncommon Maneuver - Additional Enactment Exhaust Cost)
```json
{
  "id": "animus_funeral_fungus",
  "name": "Funeral Fungus",
  "cardType": "MANEUVER",
  "faction": "ANIMUS",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 1,
  "additionalDeploymentCosts": [
    {
      "costType": "EXHAUST_ASSET",
      "filter": { "allegiance": "ALLIED", "entityType": "UNIT", "readiness": "READIED" }
    }
  ],
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "UNCOMMON",
  "rulesText": "As an additional cost to enact this maneuver, exhaust a readied allied unit.\nReturn chosen unit card from your discard pile to your hand.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
        "count": 1
      },
      "effect": {
        "primitive": "RETURN_TO_HAND",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
      }
    }
  ]
}
```

#### 9. `Infested Apostle` (Animus Common Unit - Optional Non-Resource Trigger Cost)
```json
{
  "id": "animus_infested_apostle",
  "name": "Infested Apostle",
  "cardType": "UNIT",
  "faction": "ANIMUS",
  "subtypes": ["Fungus", "Mage"],
  "requisitionCost": 4,
  "tacticsCost": 0,
  "power": 4,
  "health": 5,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "COMMON",
  "rulesText": "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit.",
  "abilities": [
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_DEPLOY",
      "condition": { "source": "SELF" },
      "optional": true,
      "optionalCosts": [
        { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "UNIT" } }
      ],
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "OPPOSING", "entityType": "UNIT" },
        "count": 1,
        "bindingPolicy": "UPFRONT_LOCK"
      },
      "effect": {
        "primitive": "DEAL_DAMAGE",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "amount": { "type": "STATIC", "value": 2 }
      }
    }
  ]
}
```

#### 10. `Soul Whip` (Concordia Uncommon Maneuver - Delegated Damage Source)
```json
{
  "id": "concordia_soul_whip",
  "name": "Soul Whip",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 2,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "UNCOMMON",
  "rulesText": "Chosen allied unit deals damage equal to its power to chosen opposing unit.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": [
        { "selector": "CHOSEN", "filter": { "allegiance": "ALLIED", "entityType": "UNIT" }, "count": 1 },
        { "selector": "CHOSEN", "filter": { "allegiance": "OPPOSING", "entityType": "UNIT" }, "count": 1 }
      ],
      "effect": {
        "primitive": "DEAL_DAMAGE",
        "damageSource": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "target": { "selector": "BOUND_TARGET", "targetIndex": 1 },
        "amount": {
          "type": "DYNAMIC",
          "source": "TARGET",
          "property": "COMBAT_POWER"
        }
      }
    }
  ]
}
```

#### 11. `Allshan's Battle Trance` (Path Mythic Maneuver - Forced Free-Deploy Destination Zone)
```json
{
  "id": "path_allshans_battle_trance",
  "name": "Allshan's Battle Trance",
  "cardType": "MANEUVER",
  "faction": "PATH",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 4,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "MYTHIC",
  "rulesText": "Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
        "count": 1
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "DEPLOY_FREE",
            "sourceZone": "DISCARD",
            "destinationZone": "MAIN_THEATER",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
          },
          {
            "primitive": "GRANT_KEYWORD",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "keyword": "STALWART",
            "duration": "END_OF_TURN"
          },
          {
            "primitive": "MODIFY_STAT",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "attribute": "COMBAT_POWER",
            "amount": { "type": "STATIC", "value": 2 },
            "duration": "END_OF_TURN"
          }
        ]
      }
    }
  ]
}
```

#### 12. `Consuming Earth` (Concordia Rare Maneuver - Dynamic Query Stat Comparison Filter)
```json
{
  "id": "concordia_consuming_earth",
  "name": "Consuming Earth",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "RARE",
  "rulesText": "Kill chosen opposing unit with power less than or equal to the total Requisition yield of all structures you control.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "OPPOSING",
          "entityType": "UNIT",
          "statComparison": {
            "property": "COMBAT_POWER",
            "operator": "LTE",
            "value": {
              "type": "QUERY",
              "source": "GAME_STATE",
              "query": {
                "zone": "BATTLEFIELD",
                "filter": { "allegiance": "ALLIED", "entityType": "STRUCTURE" },
                "aggregate": "SUM_ATTRIBUTE",
                "targetAttribute": "REQUISITION_YIELD"
              }
            }
          }
        },
        "count": 1
      },
      "effect": {
        "primitive": "KILL",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
      }
    }
  ]
}
```

#### 13. `Tempest Runner` (Concordia Uncommon Unit - `excludeSelf` Targeting Filter)
```json
{
  "id": "concordia_tempest_runner",
  "name": "Tempest Runner",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Cavalry", "Tempest"],
  "requisitionCost": 3,
  "tacticsCost": 0,
  "power": 3,
  "health": 2,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER", "RESERVES"],
  "rarity": "UNCOMMON",
  "rulesText": "Exhaust: Regroup another chosen allied unit in the same zone.",
  "abilities": [
    {
      "abilityType": "ACTIVATED",
      "costs": [
        { "costType": "EXHAUST_SELF" }
      ],
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "ALLIED",
          "entityType": "UNIT",
          "zone": "SAME_ZONE",
          "readiness": "EXHAUSTED",
          "excludeSelf": true
        },
        "count": 1
      },
      "effect": {
        "primitive": "REGROUP",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "count": { "type": "STATIC", "value": 1 }
      }
    }
  ]
}
```

#### 14. `Mutation Swell` (Animus Rare Maneuver - Discrete Quantifier `allowedCounts` & Causality Divisor)
```json
{
  "id": "animus_mutation_swell",
  "name": "Mutation Swell",
  "cardType": "MANEUVER",
  "faction": "ANIMUS",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "RARE",
  "rulesText": "Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "ALLIED",
          "entityType": "UNIT",
          "subtypes": ["Scratch Company"]
        },
        "allowedCounts": [2, 4]
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "KILL",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
          },
          {
            "primitive": "DEPLOY_FREE",
            "sourceZone": "DISCARD",
            "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
            "count": {
              "type": "DYNAMIC",
              "source": "CAUSALITY_CONTEXT",
              "property": "KILLED_COUNT",
              "divisor": 2
            }
          }
        ]
      }
    }
  ]
}
```
