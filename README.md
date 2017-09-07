# Metering with Google Datastore

This plugin implements [emitter.io](https://emitter.io) `Metering` interface. The usage statistics are written to Google Datastore as descendants of `contract` kind, in `MM/YYYY` format. The counters are incremented using Google Datastore transactions which allows multiple brokers increment counters concurrently. Note that this is intended for demonstration purposes and has not yet been tested in production.

## Usage

In order to use this plugin, you'd need to first build it (`go build -buildmode=plugin`) after this, launch the broker and point `metering` provider to the `.so` file generated, as in the example below.

```
{
	"listen": ":8080",
	"license": "....",
	"cluster": {
		"name": "00:00:00:00:00:01",
		"listen": ":4000",
		"advertise": "127.0.0.1:4000",
		"passphrase": "emitter-io"
	},
	"metering": {
		"provider": "/mnt/d/Workspace/Go/src/github.com/kelindar/plugins/metering/metering.so",
		"config": {
			"type": "service_account",
			"project_id": "your project",
			"private_key_id": "....",
			"private_key": "-----BEGIN PRIVATE KEY-----.....",
			"client_email": ".....@appspot.gserviceaccount.com",
			"client_id": ".....",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://accounts.google.com/o/oauth2/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/emitter-io%40appspot.gserviceaccount.com"
		}
	}
}
```

The `config` is actually Google Serivice Account credentials JSON file. The plugin will copy this in a separate `creds.json` file and provide it to the Google Cloud client internally. For more info, see: https://developers.google.com/identity/protocols/application-default-credentials.