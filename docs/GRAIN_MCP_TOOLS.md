## **Free & Starter Plan MCP Tools**

- **`myself`** : Get your Grain account information
    
    **Parameters**: None
    
- **`list_meetings / list_attended_meetings`:** Get filtered list of all accessible meetings
    
    **Key Parameters**:
    
    - `filters` (optional): Object containing:
        - `companies`: Array of company IDs (use `search_companies` first)
        - `persons`: Array of person IDs (use `search_persons` first)
        - `after_datetime`: ISO 8601 formatted datetime (e.g., "2024-01-01T00:00:00Z")
        - `before_datetime`: ISO 8601 formatted datetime
        - `participant_scope`: "internal" | "external"
        - `title_search`: Substring to match in meeting titles
    - `limit`: Results per page (1-20, default 10)
    - `cursor`: For pagination
- **`fetch_meeting`:** Get detailed information about a specific meeting
    
    **Key Parameters**:
    
    - `meeting_id`: UUID of the meeting
- **`fetch_meeting_transcript`**: Retrieve full meeting transcript
    
    **Key Parameters**:
    
    - `meeting_id`: UUID of the meeting
    - `include_timestamps`: Boolean (default: false)
- **`fetch_meeting_notes`:** Get AI-generated meeting notes (more concise than transcripts)
    
    **Parameters**:
    
    - `meeting_id`: UUID of the meeting
- **`search_meetings`:** Semantic search across all meeting transcripts
    
    **Key Parameters**:
    
    - `search_string`: Your search query (longer strings = more focused results)
    - `filters`: Same meeting filters as `list_meetings`
    - `limit`: Results to return (1-50, default 10)
    - `speaker_scope`: "all" | "internal" | "external" (default: all)
- **`search_companies`:** Find companies that participated in meetings when filtering
    
    **Key Parameters**:
    
    - `search_string`: Company name or domain
    - `filters`: Standard meeting filters
    - `limit`: Results to return (1-20, default 10)
- **`search_persons`** : Search for meeting participants when filtering
    - `search_string`: Person's name or email
    - `filters`: Standard meeting filters
    - `limit`: Results to return (1-20, default 10)
- **`list_workspace_users`** : Get all users in your Grain workspace
    
    **Parameters**: None
    

## Business & Enterprise Plan MCP Tools

- **`list_coaching_feedback`** : Get AI-generated sales coaching insights
    
    **Key Parameters**:
    
    - `filters.has_coaching_opportunities`: Set to `true` for flagged opportunities only
    - Standard meeting filters apply
- **`fetch_meeting_coaching_feedback` :** Get detailed coaching scorecard for specific meeting
    
    **Parameters**:
    
    - `meeting_id`: UUID of the meeting
- **`list_open_deals` / `list_all_deals` :** Access HubSpot Deal intelligence
    
    **Key Parameters**:
    
    - `filters` (optional): Object containing:
        - `companies`: Array of company IDs (UUIDs)
        - `deal_at_risk`: Boolean to filter at-risk deals
        - `deal_owner`: Person ID (UUID) of deal owner
        - `pipeline`: HubSpot pipeline ID
        - `title_search`: Substring to match in deal titles
    - `limit` (optional): Results per page (1-20, default 10)
    - `cursor` (optional): For pagination
- **`fetch_deal`** : Get detailed deal information
    
    **Parameters**:
    
    - `deal_id`: UUID of the deal

---

## Basic Usage Examples

Start with these foundational prompts to explore your meeting data.

**Getting Started (First Steps)**

- `Show me all action items I'm responsible for from the past 7 days`
- `Help me draft all follow-up emails for my external meetings today`

**Customer Intelligence**

- `Summarize customer mentions of pricing concerns in the last 30 days`
- `Help me understand how customers are comparing us with competitors`

**People & Company Insights**

- `Who from {Company Name} have we met with and what topics did we cover?`
- `Show me all users who mentioned feature requests or product feedback recently`

**Deal & Business Analysis**

- `What objections came up in external sales calls this week?`
- `Show me deals that haven't had meetings in the last 2 weeks`

**Content Discovery**

- `Search for discussions about our new product features`
- `What were the key decisions made in this week's internal meetings?`

---

## Structured Reporting Examples

**Using MCP Prompts**

While MCP Tools are designed to work with any natural language inputs into Chat, [Prompts](https://modelcontextprotocol.io/docs/concepts/prompts) are manually selected/triggered by the user to chain multiple tool calls together in a structured and repeatable way. The result is a set of comprehensive analysis workflows that would take hours manually but complete in minutes. We have created a initial library of 9 officially supported Prompts accessible directly within Claude.ai, you can riff on our templates to create your own prompts by copy/pasting your own changes to the prompt’s instructions.

**How to Use:**

1. **Click the + Button** in the bottom left of the Claude chat window
2. **Click "Add from Grain"** in the options list and choose a report to run
3. **Submit Prompt to Claude** with any (optional) custom instructions in the chat window
4. **Input Variables** when prompted by Claude as needed

![Screenshot 2025-06-16 at 5.25.10 PM.png](attachment:49154219-4692-467e-b2ff-00d7be62b6c9:Screenshot_2025-06-16_at_5.25.10_PM.png)
