# basicToOauth

> 🪦 **Project status: sunset.** `basicToOauth` did one job — swap a Basic auth header for an OAuth 2.0 Bearer token so legacy clients could keep reaching Exchange Web Services (EWS). Microsoft is retiring EWS in Exchange Online (access blocked from **October 2026**, fully disabled by **April 2027**), which retires this project's reason to exist right along with it. It keeps working until then — and with anything else that accepts an OAuth Bearer token — but no further development is planned. Thanks to everyone who ran it; see the EWS retirement notice below for the details and how to buy a little time.

## HTTP proxy service that transforms a basic authorisation header to an OAuth 2.0 Bearer token.

- Designed for Exchange Web Services (EWS) but it may work also with other services that require OAuth 2.0 Bearer token.
- This application is for HTTP protocol only (not SMTP, POP3, IMAP).

From 01.10.2022 the basic authentication will be deprecated by Microsoft for many services.
This package provides a simple way to migrate from basic authentication to OAuth by creating a proxy service.

- Application gets basic header and transform it to OAuth header. Rest of the request is passed to the target service unchanged.
- Application has been created mainly for Exchange Web Services (EWS) but it should work also with other services.

## ⚠️ Heads-up: Exchange Online is retiring EWS

Microsoft is **retiring Exchange Web Services (EWS) in Exchange Online**. Access starts being blocked in **October 2026** and EWS is **fully disabled by April 2027** — and enforcement is rolling out gradually, so some tenants may be cut off earlier. (On-premises **Exchange Server is not affected**.)

This proxy only swaps a Basic auth header for an OAuth token — it can't keep an API alive once Microsoft switches it off. So for Exchange Online, `basicToOauth` is living on borrowed time along with EWS itself.

**Buying time:** until the cut-off, a tenant admin can keep EWS switched on via Exchange Online PowerShell:

```powershell
Install-Module -Name ExchangeOnlineManagement
Connect-ExchangeOnline

# Organization (tenant) level - must be True (or unset) for EWS to work
Set-OrganizationConfig -EwsEnabled $true

# Per-mailbox - enable for all users
Get-CASMailbox -ResultSize Unlimited | Set-CASMailbox -EwsEnabled $true
```

**More info:**

- Microsoft — [Retirement of Exchange Web Services in Exchange Online](https://techcommunity.microsoft.com/blog/exchange/retirement-of-exchange-web-services-in-exchange-online/3924440)
- Synology — [EWS API error / 403 during Microsoft 365 backups or restores](https://kb.synology.com/en-us/APM/tutorial/Troubleshooting_EWS_migration)

## Need it on Graph? Let's talk — scoped and paid

This proxy supports 100% of EWS by understanding 0% of it — it just swaps the auth header and forwards the bytes. That's also why it can't simply "become" a Graph tool: re-implementing the whole of EWS on top of Graph is a non-starter (Microsoft itself hasn't closed every parity gap).

A **well-scoped subset**, however, is a different story. If your company genuinely depends on this and there's no clean native path to Graph, I'm open to a **paid engagement** to build a focused EWS-subset → Graph bridge for the operations you actually use (calendar, mail sync, contacts, free/busy — whatever that turns out to be).

I've shipped this shape of work before: [azureSMTPwithOAuth](https://github.com/mmalcek/azureSMTPwithOAuth) relays legacy SMTP to Microsoft 365 entirely over the Graph API, so the OAuth / token / retry plumbing is already proven — only the protocol mapping would be new.

Interested? [Open an issue](https://github.com/mmalcek/basicToOauth/issues/new?title=Graph+migration+inquiry) outlining your client app and the EWS operations you rely on, and we'll take the details from there.

## Downloads

Grab the latest Windows or Linux (64-bit) build from the releases page:
https://github.com/mmalcek/basicToOauth/releases

Each release is built automatically by GitHub Actions from the tagged source and
ships a `*.zip` plus a matching `*.zip.sha256` checksum. **Verify before running** —
this proxy handles credentials, so only run binaries you can confirm came from here:

```sh
# Linux / macOS
sha256sum -c basicToOauth_Linux_amd64_1-0-4.zip.sha256
```

```powershell
# Windows (compare against the .sha256 contents)
CertUtil -hashfile basicToOauth_Windows_amd64_1-0-4.zip SHA256
```

Releases also carry a build-provenance attestation, verifiable with:

```sh
gh attestation verify <zip> --repo mmalcek/basicToOauth
```

Note: only Windows and Linux (64-bit) prebuilt binaries are published. I can build other platforms on request.

### Configuration (config.yaml):

```YAML
host: "127.0.0.1" # Host of the proxy service
port: "8085" # Port of the proxy service
client_id: "yourAzureClientID" # Azure App registration client ID
tenant_id: "yourAzureTenantID" # Azure tenant ID
proxy_url: "https://outlook.office365.com" # URL of the target service
authority_url: "https://login.microsoftonline.com/" # URL of the authority service
scopes:
  - "https://outlook.office365.com/EWS.AccessAsUser.All" # Scopes for the target service
```

host 127.0.0.1 is HIGHLY RECOMMENDED because comunication between proxy service and your application is not encrypted. In other words, basicToOauth app should be on the same machine as your application.

### Installation options:

1. You can just start the application and watch communication in command line.
2. Or install as SERVICE - Open command line as administrator and run:
   - **.\basicToOauth.exe -service install**
   - **.\basicToOauth.exe -service start**
   - **.\basicToOauth.exe -service stop**
   - **.\basicToOauth.exe -service uninstall**

Once the application is running, you can use it in your application so instead "https://outlook.office365.com/..." just use "http://127.0.0.1:8085/..."

**btw: If you like this app you can buy me a coffe ;)**

<a href='https://ko-fi.com/mmalcek' target='_blank'>
	<img height='30' style='border:0px;height:40px;' src='https://az743702.vo.msecnd.net/cdn/kofi3.png?v=0' border='0' alt='Buy Me a Coffee at ko-fi.com' />
</a>
<br />
<br />

### Setup Azure "App Registration"

[MS topic - Authenticate an EWS application by using OAuth](https://learn.microsoft.com/en-us/exchange/client-developer/exchange-web-services/how-to-authenticate-an-ews-application-by-using-oauth)

Short version:

1. Azure portal -> Azure Active Directory -> App registrations -> New registration
   - Add Name (e.g. MyApp)
   - Accounts in this organizational directory only (.... - Single tenant)
   - Public client/native https://login.microsoftonline.com/common/oauth2/nativeclient
   - Register

2. Azure portal -> Azure Active Directory -> App registrations -> MyApp -> Authentication
   - Redirect URIs
     - https://login.microsoftonline.com/common/oauth2/nativeclient (should be already there)
   - Advanced settings
     - Allow public client flows -> Yes (IMPORTANT)

3. Azure portal -> Azure Active Directory -> App registrations -> MyApp -> Manifest
   - Add the following to the manifest - section "requiredResourceAccess"

```JSON
		{
			"resourceAppId": "00000002-0000-0ff1-ce00-000000000000",
			"resourceAccess": [
				{
					"id": "3b5f3d61-589b-4a3c-a359-5dd4b5ee5bd5",
					"type": "Scope"
				}
			]
		},

```

So it should looks like:

```JSON
"requiredResourceAccess": [
		{
			"resourceAppId": "00000002-0000-0ff1-ce00-000000000000",
			"resourceAccess": [
				{
					"id": "3b5f3d61-589b-4a3c-a359-5dd4b5ee5bd5",
					"type": "Scope"
				}
			]
		},
		{
			"resourceAppId": "00000003-0000-0000-c000-000000000000",
			"resourceAccess": [
				{
					"id": "e1fe6dd8-ba31-4d61-89e7-88639da4683d",
					"type": "Scope"
				}
			]
		}
	],
```

4. Azure portal -> Azure Active Directory -> App registrations -> MyApp -> Api permissions
   Click on "Grant admin consent for "yourTenantName"

5. Azure portal -> Azure Active Directory -> App registrations -> MyApp -> Overview
   - COPY "Application (client) ID" to basicToOauth app config.yaml to client_id: "YOUR_CLIENT_ID"
   - COPY "Directory (tenant) ID" to basicToOauth app config.yaml to tenant_id: "YOUR_TENANT_ID"
