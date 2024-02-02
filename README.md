# Translate-Agent

## All-in-one image

### Setting up environment

```bash
# Path where BadgerDB will store data inside the container. Do not change
export TRANSLATE_DB_BADGERDB_PATH=/data/badgerdb 

# 3rd party Translator Service
export TRANSLATE_SERVICE_TRANSLATOR=GoogleTranslate # "" (Noop), GoogleTranslate or AWSTranslate

# Only when TRANSLATOR == GoogleTranslate
export TRANSLATE_OTHER_GOOGLE_PROJECT_ID= # Google project id
export TRANSLATE_OTHER_GOOGLE_LOCATION= # Google location e.g. global
export TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY= # Path to Google account key JSON file

# Only when TRANSLATOR == AWSTranslate
export TRANSLATE_OTHER_AWS_ACCESS_KEY_ID= # AWS access key id
export TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY= # AWS secret access key
export TRANSLATE_OTHER_AWS_REGION= # AWS region e.g. eu-west-2

# Optional

# Persist data (on Host) when deleting container.
export TRANSLATE_DB_HOST_BADGERDB_PATH=translate_badgerDB

# Provide custom envoy.yaml
# https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples
export TRANSLATE_ENVOY_CONFIG_PATH= # Path to envoy.yaml
```

### Running latest all in one image

```bash
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_DB_BADGERDB_PATH \
  -e TRANSLATE_SERVICE_TRANSLATOR \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION \
  -v $TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY:/app/google_account_key.json \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY \
  -e TRANSLATE_OTHER_AWS_REGION \
  expectdigital/translate-agent-all-in-one:latest

# Add to arguments if you want to persist data on host
# -v $TRANSLATE_DB_HOST_BADGERDB_PATH:/data/badgerdb \ 

# Add to arguments if you want to use custom envoy.yaml
# -v $TRANSLATE_ENVOY_CONFIG_PATH:/app/envoy.yaml \ 
```

### Updating all in one image

Remove the old container, pull the latest image and run it again.

```bash
docker rm -f translate-all-in-one 2> /dev/null
docker pull expectdigital/translate-agent-all-in-one
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_DB_BADGERDB_PATH \
  -e TRANSLATE_SERVICE_TRANSLATOR \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION \
  -v $TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY:/app/google_account_key.json \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY \
  -e TRANSLATE_OTHER_AWS_REGION \
  expectdigital/translate-agent-all-in-one:latest

# Add to arguments if you want to persist data on host
# -v $TRANSLATE_DB_HOST_BADGERDB_PATH:/data/badgerdb \

# Add to arguments if you want to use custom envoy.yaml
# -v $TRANSLATE_ENVOY_CONFIG_PATH:/app/envoy.yaml \
```

## TypeScript client

### Installing

#### npm

```bash
npm config set @buf:registry  https://buf.build/gen/npm/v1/

npm install @buf/expectdigital_translate-agent.bufbuild_connect-es@latest
npm install @buf/expectdigital_translate-agent.bufbuild_es@latest
```

#### Other

Other package managers and their instructions can be found at [buf.build registry](https://buf.build/expectdigital/translate-agent/assets/main) under the `Node.js` tab.
Packages that are needed:

- `bufbuild/connect-es`
- `bufbuild/es`

### Usage of TypeScript client

```typescript
import { createPromiseClient } from '@bufbuild/connect';
import { createGrpcWebTransport } from '@bufbuild/connect-web';

import { TranslateService } from '@buf/expectdigital_translate-agent.bufbuild_connect-es/translate/v1/translate_connect';
import { CreateServiceRequest, Service } from '@buf/expectdigital_translate-agent.bufbuild_es/translate/v1/translate_pb';


const transport = createGrpcWebTransport({
  baseUrl: "http://localhost:8080",
});

const client = createPromiseClient(TranslateService, transport);

client.createService({
  service: new Service({ name: "test1" })
})
```
