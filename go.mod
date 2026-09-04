module github.com/fortheanprojects/voxburst-go

go 1.21

require github.com/google/uuid v1.6.0

// default base URL pointed at an unregistered domain (socialdispatch.io); credential exposure
retract [v1.0.0, v1.0.1]
