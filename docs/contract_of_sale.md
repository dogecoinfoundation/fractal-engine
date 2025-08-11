# Contract of Sale Field

The `contract_of_sale` field has been added to mint records to support Real World Asset (RWA) tokenization. This field contains legal and transactional information about the asset being tokenized.

## JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Contract of Sale",
  "type": "object",
  "required": ["contract_metadata", "legal", "parties", "asset", "tokenization"],
  "properties": {
    "contract_metadata": {
      "type": "object",
      "required": ["schema_version", "asset_category", "contract_type", "created_at"],
      "properties": {
        "schema_version": { "type": "string" },
        "asset_category": {
          "type": "string",
          "enum": ["real_estate", "art", "vehicle", "equipment", "intellectual_property", "commodity", "other"]
        },
        "contract_type": {
          "type": "string", 
          "enum": ["purchase_agreement", "consignment", "lease_to_own", "fractional_ownership"]
        },
        "created_at": { "type": "string", "format": "date-time" },
        "last_updated": { "type": "string", "format": "date-time" }
      }
    },
    "legal": {
      "type": "object",
      "required": ["jurisdiction", "governing_law", "contract_date", "effective_date"],
      "properties": {
        "jurisdiction": { "type": "string" },
        "governing_law": { "type": "string" },
        "contract_date": { "type": "string", "format": "date-time" },
        "effective_date": { "type": "string", "format": "date-time" },
        "expiration_date": { "type": "string", "format": "date-time" }
      }
    },
    "parties": {
      "type": "object",
      "required": ["seller"],
      "properties": {
        "seller": { "$ref": "#/definitions/Party" },
        "intermediaries": {
          "type": "array",
          "items": { "$ref": "#/definitions/Intermediary" }
        }
      }
    },
    "asset": {
      "type": "object",
      "required": ["type", "title", "description", "identifiers", "location", "valuation"],
      "properties": {
        "type": { "type": "string" },
        "title": { "type": "string" },
        "description": { "type": "string" },
        "identifiers": {
          "type": "object",
          "required": ["primary_id"],
          "properties": {
            "primary_id": { "type": "string" },
            "secondary_ids": { "type": "object" }
          }
        },
        "location": {
          "type": "object",
          "required": ["type"],
          "properties": {
            "type": { "type": "string", "enum": ["physical", "digital", "mobile"] },
            "address": { "type": "string" },
            "coordinates": {
              "type": "object",
              "properties": {
                "latitude": { "type": "number" },
                "longitude": { "type": "number" }
              }
            },
            "storage_requirements": { "type": "string" }
          }
        },
        "specifications": { "type": "object" },
        "condition": {
          "type": "object",
          "properties": {
            "status": { "type": "string", "enum": ["excellent", "good", "fair", "poor", "damaged"] },
            "last_assessed": { "type": "string", "format": "date-time" },
            "assessor": { "type": "string" },
            "notes": { "type": "string" }
          }
        },
        "provenance": { "type": "array" },
        "valuation": {
          "type": "object",
          "required": ["current_value", "currency", "valuation_date", "valuation_method"],
          "properties": {
            "current_value": { "type": "number" },
            "currency": { "type": "string" },
            "valuation_date": { "type": "string", "format": "date-time" },
            "valuation_method": { 
              "type": "string", 
              "enum": ["appraisal", "market_comparison", "cost_approach", "income_approach"]
            },
            "appraiser": { "type": "string" },
            "supporting_documents": { "type": "array", "items": { "type": "string" } }
          }
        }
      }
    },
    "tokenization": {
      "type": "object",
      "required": ["total_fractions", "minimum_investment", "token_rights", "compliance"],
      "properties": {
        "total_fractions": { "type": "number" },
        "minimum_investment": { "type": "number" },
        "token_rights": {
          "type": "object",
          "properties": {
            "ownership_percentage": { "type": "boolean" },
            "voting_rights": { "type": "boolean" },
            "income_distribution": { "type": "boolean" },
            "liquidation_rights": { "type": "boolean" },
            "transfer_rights": {
              "type": "object",
              "properties": {
                "freely_transferable": { "type": "boolean" },
                "restrictions": { "type": "array", "items": { "type": "string" } }
              }
            }
          }
        },
        "compliance": {
          "type": "object",
          "properties": {
            "securities_exemption": { "type": "string" },
            "accredited_only": { "type": "boolean" },
            "geographic_restrictions": { "type": "array", "items": { "type": "string" } },
            "kyc_aml_required": { "type": "boolean" }
          }
        }
      }
    },
    "contingencies": {
      "type": "array",
      "items": { "$ref": "#/definitions/Contingency" }
    },
    "legal_protections": { "$ref": "#/definitions/LegalProtections" },
    "ongoing_obligations": { "$ref": "#/definitions/OngoingObligations" },
    "dispute_resolution": { "$ref": "#/definitions/DisputeResolution" },
    "additional_clauses": { "$ref": "#/definitions/AdditionalClauses" }
  },
  "definitions": {
    "Party": {
      "type": "object",
      "required": ["name", "entity_type", "address"],
      "properties": {
        "name": { "type": "string" },
        "entity_type": { 
          "type": "string", 
          "enum": ["individual", "corporation", "llc", "partnership", "trust", "dao"]
        },
        "address": { "type": "string" },
        "identification": {
          "type": "object",
          "properties": {
            "tax_id": { "type": "string" },
            "registration_number": { "type": "string" },
            "kyc_status": { "type": "string", "enum": ["verified", "pending", "required"] }
          }
        },
        "contact": {
          "type": "object",
          "properties": {
            "email": { "type": "string", "format": "email" },
            "phone": { "type": "string" }
          }
        }
      }
    },
    "Intermediary": {
      "type": "object",
      "required": ["role", "name"],
      "properties": {
        "role": { 
          "type": "string", 
          "enum": ["broker", "escrow", "custodian", "appraiser", "inspector"]
        },
        "name": { "type": "string" },
        "license_number": { "type": "string" }
      }
    },
    "Contingency": {
      "type": "object",
      "required": ["type", "description", "satisfied"],
      "properties": {
        "type": { 
          "type": "string", 
          "enum": ["inspection", "appraisal", "financing", "title", "insurance", "regulatory"]
        },
        "description": { "type": "string" },
        "deadline": { "type": "string", "format": "date-time" },
        "satisfied": { "type": "boolean" },
        "conditions": { "type": "object" }
      }
    },
    "LegalProtections": {
      "type": "object",
      "properties": {
        "force_majeure": {
          "type": "object",
          "properties": {
            "included": { "type": "boolean" },
            "events": { "type": "array", "items": { "type": "string" } }
          }
        },
        "warranties": { "type": "array", "items": { "type": "string" } },
        "indemnification": {
          "type": "object",
          "properties": {
            "scope": { "type": "array", "items": { "type": "string" } },
            "time_limit_years": { "type": "number" },
            "monetary_cap": { "type": "number" }
          }
        },
        "insurance_requirements": { "type": "array", "items": { "type": "string" } }
      }
    },
    "OngoingObligations": {
      "type": "object",
      "properties": {
        "maintenance": { "type": "string" },
        "insurance": { "type": "array", "items": { "type": "string" } },
        "reporting": {
          "type": "object",
          "properties": {
            "frequency": { "type": "string", "enum": ["monthly", "quarterly", "annually"] },
            "requirements": { "type": "array", "items": { "type": "string" } }
          }
        },
        "compliance": { "type": "array", "items": { "type": "string" } },
        "expense_allocation": { "type": "string" }
      }
    },
    "DisputeResolution": {
      "type": "object",
      "properties": {
        "governing_law": { "type": "string" },
        "jurisdiction": { "type": "string" },
        "arbitration_required": { "type": "boolean" },
        "arbitration_rules": { "type": "string" },
        "mediation_first": { "type": "boolean" }
      }
    },
    "AdditionalClauses": {
      "type": "object",
      "properties": {
        "asset_specific": { "type": "object" },
        "custom_provisions": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "title": { "type": "string" },
              "description": { "type": "string" },
              "text": { "type": "string" }
            }
          }
        },
        "special_conditions": { "type": "array", "items": { "type": "string" } },
        "regulatory_notes": { "type": "string" }
      }
    }
  }
}
```

## Structure

The `contract_of_sale` field is a JSON object that can contain the following information:

### Contract Metadata
```json
{
  "contract_metadata": {
    "schema_version": "1.0",
    "asset_category": "real_estate",
    "contract_type": "fractional_ownership",
    "created_at": "2024-01-15T10:30:00Z",
    "last_updated": "2024-01-15T10:30:00Z"
  }
}
```

### Legal Information
```json
{
  "legal": {
    "jurisdiction": "Delaware, USA",
    "governing_law": "Delaware State Law",
    "contract_date": "2024-01-15T10:30:00Z",
    "effective_date": "2024-01-20T00:00:00Z",
    "expiration_date": "2025-01-20T00:00:00Z"
  }
}
```

### Parties Information
```json
{
  "parties": {
    "seller": {
      "name": "ABC Holdings LLC",
      "address": "123 Main St, Wilmington, DE 19801",
      "entity_type": "llc",
      "identification": {
        "tax_id": "12-3456789",
        "registration_number": "DE-LLC-123456",
        "kyc_status": "verified"
      },
      "contact": {
        "email": "contact@abcholdings.com",
        "phone": "+1-555-123-4567"
      }
    },
    "intermediaries": [
      {
        "role": "broker",
        "name": "Premium Real Estate Services",
        "license_number": "RE-2024-001"
      }
    ]
  }
}
```

### Asset Details
```json
{
  "asset": {
    "type": "commercial_office_building",
    "title": "Downtown SF Office Complex",
    "description": "Premium Class A office building in SOMA district",
    "identifiers": {
      "primary_id": "SF-APN-123-456-789",
      "secondary_ids": {
        "building_id": "SOMA-TOWER-1",
        "leed_id": "LEED-2024-001"
      }
    },
    "location": {
      "type": "physical",
      "address": "789 Business Blvd, San Francisco, CA 94102",
      "coordinates": {
        "latitude": 37.7749,
        "longitude": -122.4194
      }
    },
    "specifications": {
      // See asset-specific examples below
    },
    "condition": {
      "status": "excellent",
      "last_assessed": "2024-01-10T00:00:00Z",
      "assessor": "Professional Building Inspectors Inc",
      "notes": "Well-maintained property with recent upgrades"
    },
    "valuation": {
      "current_value": 15000000,
      "currency": "USD",
      "valuation_date": "2024-01-10T00:00:00Z",
      "valuation_method": "income_approach",
      "appraiser": "Commercial Property Valuers LLC"
    }
  }
}
```

### Tokenization
```json
{
  "tokenization": {
    "total_fractions": 1000000,
    "minimum_investment": 1000,
    "token_rights": {
      "ownership_percentage": true,
      "voting_rights": true,
      "income_distribution": true,
      "liquidation_rights": true,
      "transfer_rights": {
        "freely_transferable": false,
        "restrictions": ["accredited_investors_only", "lock_period_365_days"]
      }
    },
    "compliance": {
      "securities_exemption": "regulation_d",
      "accredited_only": true,
      "geographic_restrictions": ["US", "EU"],
      "kyc_aml_required": true
    }
  }
}
```

### Verification and Compliance
```json
{
  "verification": {
    "title_clear": true,
    "liens": [],
    "encumbrances": [],
    "permits_current": true,
    "zoning_compliant": true,
    "environmental_clearance": true
  },
  "compliance": {
    "aml_kyc_completed": true,
    "accredited_investor_verified": true,
    "regulatory_approvals": ["SEC_REGULATION_D", "STATE_SECURITIES_EXEMPTION"]
  }
}
```

### Documentation
```json
{
  "documents": {
    "contract_hash": "sha256:abc123...",
    "deed_hash": "sha256:def456...",
    "appraisal_hash": "sha256:ghi789...",
    "title_insurance_hash": "sha256:jkl012...",
    "inspection_reports": ["sha256:mno345...", "sha256:pqr678..."]
  }
}
```

### Rights and Obligations
```json
{
  "rights": {
    "ownership_percentage": 100,
    "voting_rights": true,
    "income_distribution": true,
    "liquidation_rights": true,
    "information_rights": ["financial_statements", "operating_reports", "material_changes"],
    "transfer_restrictions": {
      "lock_period_days": 365,
      "accredited_only": true,
      "geographic_restrictions": ["US", "EU"],
      "right_of_first_refusal": true
    }
  },
  "obligations": {
    "maintenance_responsibility": "token_holders",
    "insurance_requirements": {
      "minimum_coverage": 15000000,
      "types": ["property", "liability", "business_interruption"]
    },
    "reporting_frequency": "quarterly",
    "audit_requirements": "annual_independent_audit",
    "expense_allocation": "pro_rata_by_ownership"
  }
}
```

### Contingencies and Conditions
```json
{
  "contingencies": [
    {
      "type": "inspection",
      "description": "Professional building inspection",
      "deadline": "2024-02-01T17:00:00Z",
      "satisfied": false,
      "conditions": {
        "professional_inspection": true,
        "structural_assessment": true
      }
    },
    {
      "type": "title",
      "description": "Clear title verification",
      "deadline": "2024-02-10T17:00:00Z",
      "satisfied": false,
      "conditions": {
        "clear_title_required": true,
        "lien_search_complete": true
      }
    }
  ]
}
```

### Legal Protections and Disclaimers
```json
{
  "legal_protections": {
    "force_majeure": {
      "included": true,
      "events": ["natural_disasters", "government_action", "war", "pandemic", "cyber_attacks"]
    },
    "material_adverse_change": {
      "clause_included": true,
      "scope": "asset_value_decrease_exceeding_10_percent"
    },
    "indemnification": {
      "seller_indemnifies": true,
      "scope": ["title_defects", "environmental_issues", "undisclosed_liabilities"],
      "time_limit_years": 3,
      "monetary_cap": 1500000
    },
    "warranty_disclaimers": {
      "as_is_clause": false,
      "seller_warranties": ["clear_title", "no_material_defects", "compliance_with_laws"]
    }
  }
}
```

### Dispute Resolution
```json
{
  "dispute_resolution": {
    "governing_law": "Delaware State Law",
    "jurisdiction": "Delaware Courts",
    "arbitration": {
      "required": true,
      "organization": "American Arbitration Association",
      "rules": "Commercial Arbitration Rules",
      "location": "Wilmington, Delaware"
    },
    "mediation_first": true,
    "attorney_fees": "prevailing_party_recovers"
  }
}
```

### Additional Clauses
```json
{
  "additional_clauses": {
    "asset_specific": {
      "property_management": "professional_management_required",
      "tenant_approval": "major_leases_require_token_holder_approval"
    },
    "custom_provisions": [
      {
        "title": "Technology Integration Clause",
        "description": "Buyer shall have the right to install IoT sensors and smart building technology",
        "text": "The token holders may, at their own expense, install Internet of Things (IoT) sensors, smart meters, and building automation systems to monitor and optimize the Property's performance, provided such installations do not materially alter the structure or violate applicable building codes."
      }
    ],
    "special_conditions": [
      "Property must maintain LEED Gold certification",
      "Existing tenant leases to be assigned without modification"
    ],
    "regulatory_compliance": {
      "securities_law_compliance": "Regulation D, Rule 506(c)",
      "aml_kyc_requirements": "Full verification required for all token holders",
      "ongoing_reporting": "Quarterly reports to token holders, annual SEC filing"
    }
  }
}
```

## Asset-Specific Specification Examples

The `specifications` field within the `asset` object should be customized based on the asset type:

### Real Estate Specifications
```json
"specifications": {
  "property_type": "commercial",
  "square_footage": 50000,
  "lot_size": 100000,
  "year_built": 2015,
  "zoning": "commercial_office",
  "occupancy": {
    "current_occupancy_rate": 0.95,
    "tenant_count": 25,
    "average_lease_term": 60
  },
  "income": {
    "annual_rental_income": 1200000,
    "operating_expenses": 400000,
    "net_operating_income": 800000
  },
  "features": ["parking_garage", "elevator", "hvac_2020", "fiber_optic"],
  "certifications": ["leed_gold", "energy_star"]
}
```

### Art Specifications
```json
"specifications": {
  "artist": "Pablo Picasso",
  "creation_year": 1907,
  "medium": "oil_on_canvas",
  "dimensions": {
    "height_cm": 243.9,
    "width_cm": 233.7,
    "depth_cm": 5.0
  },
  "style": "cubism",
  "period": "african_period",
  "edition": {
    "type": "original",
    "number": "1/1",
    "total_editions": 1
  },
  "authentication": {
    "certificate_number": "AUTH-2024-001",
    "issuing_authority": "Picasso Authentication Board",
    "authentication_date": "2024-01-05T00:00:00Z"
  },
  "exhibition_history": [
    {
      "venue": "Museum of Modern Art",
      "exhibition": "Picasso Retrospective",
      "dates": "2023-03-01 to 2023-06-01"
    }
  ],
  "conservation": {
    "last_conservation": "2023-01-15T00:00:00Z",
    "conservator": "Metropolitan Museum Conservation Lab",
    "condition_report": "Excellent, minor frame wear"
  }
}
```

### Vehicle Specifications
```json
"specifications": {
  "make": "Ferrari",
  "model": "250 GTO",
  "year": 1962,
  "vin": "3851GT",
  "engine": {
    "type": "V12",
    "displacement_liters": 3.0,
    "horsepower": 300
  },
  "transmission": "5_speed_manual",
  "mileage": 45000,
  "color": {
    "exterior": "Rosso Corsa",
    "interior": "Black Leather"
  },
  "features": ["numbers_matching", "original_engine", "competition_history"],
  "history": {
    "racing_history": [
      {
        "event": "Le Mans 24 Hours",
        "year": 1962,
        "result": "2nd_place"
      }
    ],
    "accident_history": [],
    "service_records": true,
    "previous_owners": 3
  },
  "documentation": {
    "title_clear": true,
    "registration_current": true,
    "ferrari_classiche_certified": true
  }
}
```

### Equipment/Machinery Specifications
```json
"specifications": {
  "manufacturer": "Caterpillar",
  "model": "D11T",
  "serial_number": "CAT123456789",
  "year_manufactured": 2020,
  "equipment_type": "bulldozer",
  "operating_hours": 2500,
  "condition_rating": "excellent",
  "maintenance": {
    "last_service": "2024-01-01T00:00:00Z",
    "service_interval_hours": 500,
    "warranty_remaining": true,
    "warranty_expires": "2025-01-01T00:00:00Z"
  },
  "specifications_technical": {
    "operating_weight_kg": 104326,
    "engine_power_hp": 850,
    "blade_capacity_m3": 19.0
  },
  "location_usage": {
    "current_site": "Colorado Mining Operation",
    "usage_type": "mining",
    "operator_certified": true
  }
}
```

### Intellectual Property Specifications
```json
"specifications": {
  "ip_type": "patent",
  "title": "Method for Cryptocurrency Transaction Verification",
  "registration_number": "US11234567B2",
  "filing_date": "2022-01-15T00:00:00Z",
  "grant_date": "2024-01-15T00:00:00Z",
  "expiration_date": "2042-01-15T00:00:00Z",
  "jurisdiction": ["United States", "European Union", "Japan"],
  "claims": 20,
  "prior_art_references": 15,
  "licensing": {
    "current_licenses": [
      {
        "licensee": "Tech Corp Inc",
        "territory": "North America",
        "royalty_rate": 0.05,
        "annual_revenue": 50000
      }
    ],
    "licensing_potential": "high"
  },
  "enforcement": {
    "litigation_history": [],
    "opposition_proceedings": []
  }
}
```

## Example Complete Contract of Sale

```json
{
  "contract_metadata": {
    "schema_version": "1.0",
    "asset_category": "real_estate",
    "contract_type": "fractional_ownership",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "legal": {
    "jurisdiction": "California, USA",
    "governing_law": "California Commercial Code",
    "contract_date": "2024-01-15T10:30:00Z",
    "effective_date": "2024-01-20T00:00:00Z"
  },
  "parties": {
    "seller": {
      "name": "SF Commercial Properties LLC",
      "address": "100 Market St, San Francisco, CA 94105",
      "entity_type": "llc",
      "identification": {
        "tax_id": "12-3456789",
        "registration_number": "CA-LLC-789012",
        "kyc_status": "verified"
      }
    }
  },
  "asset": {
    "type": "commercial_real_estate",
    "title": "Class A office building, 50,000 sq ft, built 2015",
    "description": "Premium office space in downtown San Francisco",
    "identifiers": {
      "primary_id": "SF-APN-123-456-789"
    },
    "location": {
      "type": "physical",
      "address": "789 Business Blvd, San Francisco, CA 94102"
    },
    "specifications": {
      "property_type": "commercial",
      "square_footage": 50000,
      "year_built": 2015,
      "occupancy": {
        "current_occupancy_rate": 0.95,
        "tenant_count": 25
      },
      "income": {
        "annual_rental_income": 1200000,
        "operating_expenses": 400000,
        "net_operating_income": 800000
      }
    },
    "valuation": {
      "current_value": 15000000,
      "currency": "USD",
      "valuation_date": "2024-01-10T00:00:00Z",
      "valuation_method": "income_approach"
    }
  },
  "tokenization": {
    "total_fractions": 1000000,
    "minimum_investment": 1000,
    "token_rights": {
      "ownership_percentage": true,
      "voting_rights": true,
      "income_distribution": true,
      "liquidation_rights": true,
      "transfer_rights": {
        "freely_transferable": false,
        "restrictions": ["accredited_investors_only", "lock_period_365_days"]
      }
    },
    "compliance": {
      "securities_exemption": "regulation_d",
      "accredited_only": true,
      "geographic_restrictions": ["US"],
      "kyc_aml_required": true
    }
  },
  "contingencies": [
    {
      "type": "inspection",
      "description": "Professional building inspection required",
      "deadline": "2024-02-01T17:00:00Z",
      "satisfied": false,
      "conditions": {
        "professional_inspection": true
      }
    },
    {
      "type": "title",
      "description": "Clear title verification required",
      "deadline": "2024-02-10T17:00:00Z",
      "satisfied": false,
      "conditions": {
        "clear_title_required": true
      }
    }
  ],
  "legal_protections": {
    "force_majeure": {
      "included": true,
      "events": ["natural_disasters", "government_action", "pandemic"]
    },
    "indemnification": {
      "scope": ["title_defects", "environmental_issues"],
      "time_limit_years": 2,
      "monetary_cap": 500000
    }
  },
  "additional_clauses": {
    "special_conditions": [
      "Property must maintain current tenant occupancy rate above 90%",
      "Seller to provide 6-month property management transition"
    ],
    "regulatory_compliance": {
      "securities_law_compliance": "Regulation D, Rule 506(c)",
      "aml_kyc_requirements": "Full verification required for all token holders"
    }
  }
}
```

## Usage in Minting

When creating a mint for an RWA, include the `contract_of_sale` field in your mint request:

```json
{
  "title": "SF Commercial Building Token",
  "description": "Fractional ownership of premium SF office building",
  "fraction_count": 1000000,
  "contract_of_sale": {
    // ... contract of sale data as shown above
  }
  // ... other mint fields
}
```

The `contract_of_sale` field is included in the mint hash calculation, ensuring the legal contract details are immutably tied to the tokenized asset.

## Benefits

This structure provides:
- **Flexibility** - Works across all asset types
- **Extensibility** - Easy to add new asset categories and specifications
- **Validation** - JSON Schema ensures data integrity
- **Consistency** - Common legal/tokenization framework
- **Specificity** - Asset-specific details where needed
- **Compliance** - Built-in regulatory considerations
