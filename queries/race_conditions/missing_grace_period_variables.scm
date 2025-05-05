; Name: Missing Grace Period Variables
; Description: Contract doesn't have storage for grace period enforcement
; Severity: Medium
; References: Front-running protection pattern

; Find storage struct without grace period or timelock variables
(struct_item
  name: (type_identifier) @struct_name
  (#match? @struct_name "Storage")
  body: (field_declaration_list) @field_list
  (#not-match? @field_list "grace_period|timelock|delay"))
