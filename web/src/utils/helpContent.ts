export const helpContent = {
  dashboard: {
    overview:
      "The dashboard provides a real-time overview of your mail server's status and recent activity. Metrics update automatically every 30 seconds.",
    queueMessages:
      "Total number of messages currently in the mail queue waiting for delivery.",
    deliveredToday:
      "Number of messages successfully delivered today. Success rate shows the percentage of successful deliveries.",
    deferred:
      "Messages that failed delivery temporarily and are scheduled for retry. These are usually due to temporary issues like DNS problems or recipient server unavailability.",
    frozen:
      "Messages that have been paused and won't be retried automatically. These require manual intervention to resume delivery.",
    failedToday:
      "Messages that failed permanently today due to issues like invalid recipients or rejected content.",
    pendingToday:
      "Messages that are currently being processed or waiting for delivery attempts.",
    logEntries: "Number of log entries processed today from Exim log files.",
    weeklyChart:
      "Shows email activity over the past 7 days. Green bars represent delivered messages, red bars show failures, yellow shows pending messages, and gray shows deferred messages.",
  },

  queue: {
    overview:
      "The queue management interface allows you to view, search, and manage all messages currently in the mail queue.",
    messageId:
      "Unique identifier assigned by Exim to each message. Used for tracking and operations.",
    sender:
      "The envelope sender address - the address that will receive bounce notifications.",
    recipients:
      "List of recipient addresses for this message. Multi-recipient messages show the first recipient with a count.",
    size: "Message size in bytes, including headers and content.",
    age: "How long the message has been in the queue since it was first received.",
    status: {
      queued:
        "Message is in the queue and will be delivered according to normal retry schedule.",
      deferred:
        "Message failed delivery temporarily and is scheduled for retry later.",
      frozen:
        "Message has been paused and won't be retried until manually thawed.",
    },
    retryCount: "Number of delivery attempts made for this message.",
    operations: {
      deliver:
        "Force immediate delivery attempt, bypassing the normal retry schedule. Useful for messages that failed due to temporary issues that have been resolved.",
      freeze:
        "Pause the message to prevent further delivery attempts. Use this for messages that need investigation or are causing problems.",
      thaw: "Resume a frozen message and return it to normal retry scheduling.",
      delete:
        "Permanently remove the message from the queue. This action cannot be undone and the message will be lost.",
    },
    bulkOperations:
      "Select multiple messages using checkboxes to perform operations on many messages at once. Use with caution as bulk operations affect all selected messages.",
    search: {
      overview:
        "Use advanced search to filter messages by various criteria. Multiple filters can be combined.",
      sender:
        "Filter by sender email address. Use wildcards like *@domain.com to match all senders from a domain.",
      recipient:
        "Filter by recipient email address. Supports wildcards and partial matches.",
      messageId: "Search for a specific message by its unique ID.",
      age: "Filter by message age. Examples: >2h (older than 2 hours), <30m (newer than 30 minutes), =1d (exactly 1 day old).",
      status: "Filter by message status: queued, deferred, or frozen.",
      retryCount:
        "Filter by number of delivery attempts. Examples: >3 (more than 3 attempts), =0 (no attempts yet).",
    },
  },

  logs: {
    overview:
      "The log viewer provides access to Exim log files with powerful search and filtering capabilities.",
    logTypes: {
      main: "Contains message arrivals, deliveries, deferrals, and bounces. This is the primary log for tracking message flow.",
      reject:
        "Contains rejected messages and connections, including spam filtering and policy rejections.",
      panic:
        "Contains daemon errors and critical failures that require immediate attention.",
    },
    events: {
      arrival: "Message was received and accepted into the queue.",
      delivery: "Message was successfully delivered to a recipient.",
      defer: "Delivery attempt failed temporarily and will be retried.",
      bounce: "Delivery failed permanently and a bounce message was generated.",
      reject: "Message or connection was rejected before being accepted.",
    },
    realTimeTail:
      "Monitor new log entries as they occur in real-time. Use filters to show only relevant entries. Click 'Pause' to stop auto-scrolling while reviewing entries.",
    search: {
      dateRange:
        "Specify start and end dates to limit search to a specific time period.",
      textSearch:
        "Search within log entry text for specific keywords or error messages.",
      messageId: "Find all log entries related to a specific message ID.",
      eventType:
        "Filter by specific event types like delivery, defer, or bounce.",
    },
    export:
      "Export filtered log entries for external analysis. Choose CSV format for spreadsheet analysis or TXT format for plain text processing.",
  },

  reports: {
    overview:
      "Comprehensive analytics and reporting on your mail server's performance and deliverability.",
    deliverability: {
      overview:
        "Shows success rates and delivery metrics over time to help identify trends and issues.",
      successRate:
        "Percentage of messages that were delivered successfully on the first or subsequent attempts.",
      deferRate:
        "Percentage of messages that failed temporarily and required retry attempts.",
      bounceRate:
        "Percentage of messages that failed permanently and could not be delivered.",
      topSenders:
        "Senders with the highest message volume, useful for identifying bulk senders and their performance.",
      topRecipients:
        "Recipient domains with the highest volume, helpful for identifying delivery patterns.",
      domainAnalysis:
        "Per-domain deliverability statistics to identify problematic destinations.",
    },
    volume: {
      overview:
        "Traffic trends and patterns to understand mail server usage and capacity planning.",
      hourlyVolume:
        "Messages per hour showing daily traffic patterns and peak usage times.",
      dailyVolume: "Messages per day showing weekly and monthly trends.",
      peakHours:
        "Busiest times of day to help with capacity planning and maintenance scheduling.",
      growthTrends:
        "Volume changes over time to identify growth patterns or unusual activity.",
    },
    failures: {
      overview:
        "Analysis of delivery failures to identify common issues and problem areas.",
      failureCategories:
        "Common failure types grouped by cause (DNS issues, policy rejections, etc.).",
      bounceCodes:
        "SMTP response codes and their frequencies to understand why messages are failing.",
      problemDomains:
        "Domains with high failure rates that may need special attention or configuration.",
      retryPatterns: "Analysis of retry behavior to optimize retry schedules.",
    },
  },

  messageTrace: {
    overview:
      "Detailed delivery history and troubleshooting information for individual messages.",
    deliveryTimeline:
      "Complete chronological history of the message from receipt to final delivery or failure.",
    recipientTracking:
      "For multi-recipient messages, shows individual status for each recipient address.",
    smtpDetails:
      "Detailed SMTP transaction information including server responses and error messages.",
    troubleshootingNotes:
      "Add notes and tags to messages for tracking investigation progress and sharing information with team members.",
    deliveryAttempts:
      "Each delivery attempt with timestamp, destination server, SMTP response, and outcome.",
    retrySchedule:
      "Shows when the next delivery attempt is scheduled based on Exim's retry configuration.",
  },

  security: {
    authentication:
      "All administrative functions require authentication. Sessions automatically expire after inactivity for security.",
    auditLogging:
      "All administrative actions are logged with user identity, timestamp, and details for compliance and security monitoring.",
    dataProtection:
      "Message content viewing includes safety measures to protect sensitive information. Personal data can be masked or redacted.",
    accessControl:
      "Different permission levels control what actions users can perform (read-only vs. administrative access).",
  },

  troubleshooting: {
    queueNotLoading:
      "If the queue doesn't load, check that Exim is running and the application has permission to access queue files.",
    logsNotUpdating:
      "If logs aren't updating, verify log file permissions and check that the log monitoring service is running.",
    performanceIssues:
      "For slow performance, try using more specific search filters, reducing page sizes, or checking system resources.",
    authenticationProblems:
      "For login issues, verify credentials, check session configuration, and ensure system time is synchronized.",
    operationsFailing:
      "If message operations fail, check that the application has permission to execute Exim commands.",
  },
};

export const getHelpContent = (
  section: string,
  subsection?: string
): string => {
  const content = helpContent as any;
  if (subsection) {
    // Handle nested properties with dot notation (e.g., 'status.queued')
    if (subsection.includes('.')) {
      const parts = subsection.split('.');
      let value = content[section];
      for (const part of parts) {
        value = value?.[part];
        if (!value) break;
      }
      return value || "Help content not available.";
    }
    return content[section]?.[subsection] || "Help content not available.";
  }
  return (
    content[section]?.overview ||
    content[section] ||
    "Help content not available."
  );
};
