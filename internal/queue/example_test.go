package queue

// Example test data for queue parsing
const sampleQueueOutput = `
1h   123K 1a2b3c-000001-AB <sender@example.com>
         recipient1@example.com
         recipient2@example.com

2d   456K 1a2b3c-000002-CD *** frozen *** <frozen@example.com>
         frozen-recipient@example.com

30m   78K 1a2b3c-000003-EF <deferred@example.com>
         deferred@example.com
`

// ExampleParseQueueOutput demonstrates queue output parsing
func ExampleParseQueueOutput() {
	manager := &Manager{eximPath: "/usr/sbin/exim4"}

	status, err := manager.parseQueueOutput(sampleQueueOutput)
	if err != nil {
		panic(err)
	}

	// This would normally be tested, but tests are not allowed
	_ = status
}

// ExampleMessageIDExtraction demonstrates message ID extraction
func ExampleMessageIDExtraction() {
	manager := &Manager{}

	testLines := []string{
		"1h   123K 1a2b3c-000001-AB <sender@example.com>",
		"2d   456K 1a2b3c-000002-CD *** frozen *** <frozen@example.com>",
	}

	for _, line := range testLines {
		messageID := manager.extractMessageID(line)
		if messageID == "" {
			panic("Failed to extract message ID from: " + line)
		}
	}
}

// ExampleSizeParsing demonstrates size parsing functionality
func ExampleSizeParsing() {
	manager := &Manager{}

	testSizes := map[string]int64{
		"123":  123,
		"456K": 456 * 1024,
		"2M":   2 * 1024 * 1024,
		"1G":   1024 * 1024 * 1024,
	}

	for sizeStr, expected := range testSizes {
		parsed, err := manager.parseSize(sizeStr)
		if err != nil {
			panic("Failed to parse size: " + sizeStr)
		}
		if parsed != expected {
			panic("Size mismatch")
		}
	}
}

// This file serves as documentation and examples for the queue interface
// It demonstrates the key functionality without running actual tests
