# basicToOauth 

## HTTP proxy service that transforms a basic authorisation header to an OAuth 2.0 Bearer token. 
- Designed for Exchange Web Services (EWS) but it may work also with other services that require OAuth 2.0 Bearer token.
- This application is for HTTP protocol only (not SMTP, POP3, IMAP). 


From 01.10.2022 the basic authentication will be deprecated by Microsoft for many services. 
This package provides a simple way to migrate from basic authentication to OAuth by creating a proxy service.

- Application gets basic header and transform it to OAuth header. Rest of the request is passed to the target service unchanged.
- Application has been created mainly for Exchange Web Services (EWS) but it should work also with other services.

## You can download Windows version from here:
https://github.com/mmalcek/basicToOauth/releases
<br />
[Direct Windows download link](https://github.com/mmalcek/basicToOauth/releases/download/v1.0.3/basicToOauth_Windows_amd64_1-0-3.zip)
<br />
Note: Currently only Windows and Linux (64bit) prebuild binaries are available. I can build binaries for other platforms on request.


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
