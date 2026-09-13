# Google Workspace and OAuth

Google Workspace requires OAuth 2.0 for CalDAV. Basic auth and app-specific
passwords return `401 Unauthorized` on Google's current CalDAV endpoint.

DankCalendar does not provide a hosted or shared Google OAuth application.
Each user supplies their own Google Cloud OAuth desktop client for their own
Google account or Workspace. If Google shows `dankcalendar has not completed
the Google verification process`, that message refers to the OAuth app in
your Google Cloud project. For a personal or test setup, add the Google
account under the OAuth consent screen's test users. For a public app,
complete Google's OAuth verification.

## Adding Google calendars

1. Create a Google OAuth desktop client ID in Google Cloud Console and enable
   the Google Calendar API.
2. Run:

   ```sh
   dankcalendar google-discover --account you@example.com --client-id YOUR_CLIENT_ID.apps.googleusercontent.com
   ```

3. Complete the browser authorization flow. DankCalendar stores the refresh
   token in the system keyring and writes discovered calendars to the normal
   config file.

Discovered Google calendars use Google's CalDAV endpoint with OAuth
bearer-token authentication. Event listing, creation, editing, and deletion
still use DankCalendar's CalDAV backend.
