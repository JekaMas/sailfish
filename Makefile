.PHONY: test vet race bench fuzz

test:
	GOWORK=off go test ./...

vet:
	GOWORK=off go vet ./...

race:
	GOWORK=off go test -race ./...

bench:
	GOWORK=off go test -run '^$$' -bench . -benchmem -count=5

fuzz:
	GOWORK=off go test -run '^$$' -fuzz '^FuzzPriceInUint64UnitsDecimalPlaces9ParseRoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzUint64UnitsRoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzNativeUnitWidthsRoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzUint256UnitsRoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzJSONRoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzCBORUint64RoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzCBORUint256RoundTrip$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzCBORDecoderAcceptsOnlyPreferredRoundTrips$$' -fuzztime=5s
	GOWORK=off go test -run '^$$' -fuzz '^FuzzCBORFirstConsumesExactlyOnePreferredValue$$' -fuzztime=5s
