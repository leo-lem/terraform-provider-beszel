# By PocketBase record id:
terraform import beszel_alert.example i4bzj802ik4qynl

# Or by the natural key Beszel enforces a unique index on (user/system/name),
# which is easier to find from the hub UI than the internal record id:
terraform import beszel_alert.example lo3vm112x037lr9/mgkg479e40815tj/CPU
