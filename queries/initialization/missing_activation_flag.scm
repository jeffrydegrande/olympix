; Name: Missing Market Activation Flag
; Description: Contract has no dedicated activation flag to prevent operation before proper initialization
; Severity: High
; References: zkLend hack (Feb 2025)

; Find storage struct without an activation flag
(struct_item
  name: (type_identifier) @struct_name
  (#match? @struct_name "Storage")
  body: (field_declaration_list) @field_list
  (#not-match? @field_list "is_active|active|isActive|marketActive"))
