# basicToOauth 

## relay service that transforms a basic authorisation header to an OAuth 2.0 Bearer token. 
- Designed for Exchange Web Services (EWS)


From 01.10.2022 the basic authentication will be deprecated by Microsoft for many services. 
This package provides a simple way to migrate from basic authentication to OAuth by creating a relay service.

- Application gets basic header and transform it to OAuth header. Rest of the request is passed to the target service unchanged.
- Application has been created mainly for Exchange Web Services (EWS) but it should work also with other services.

### Configuration (config.yaml):
```YAML
host: "127.0.0.1" # Host of the relay service
port: "8085" # Port of the relay service
client_id: "yourAzureClientID" # Azure App registration client ID
tenant_id: "yourAzureTenantID" # Azure tenant ID
proxy_url: "https://outlook.office365.com" # URL of the target service
authority_url: "https://login.microsoftonline.com/" # URL of the authority service
scopes:
  - "https://outlook.office365.com/EWS.AccessAsUser.All" # Scopes for the target service
```
host 127.0.0.1 is HIGHLY RECOMMENDED because comunication between relay service and your application is not encrypted.


### Installation options:
1. you can just start the application and watch communication in command line.
2. Install as SERVICE 
    - **.\basicToOauth -service install**
    - **.\basicToOauth -service start**
    - **.\basicToOauth -service stop**
    - **.\basicToOauth -service uninstall**

Once the application is running, you can use it in your application so instead "https://outlook.office365.com/..." just use "http://127.0.0.1:8085/..."

**If this app helped you can buy me a coffe ;)**

<a href='https://ko-fi.com/mmalcek' target='_blank'>
	<img height='25' style='border:0px;height:35px;' src='https://az743702.vo.msecnd.net/cdn/kofi3.png?v=0' border='0' alt='Buy Me a Coffee at ko-fi.com' />
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
