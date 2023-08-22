# Translate-Agent

## All-in-one image
Running latest all in one image
```bash
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID={Google project id} \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION={Google location} \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID={AWS access key id} \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY={AWS secret access key} \
  -e TRANSLATE_OTHER_AWS_REGION={AWS region} \
  -v path/to/google_account_key.json:/app/google_account_key.json \
  -v path/to/badger-dir:/data/badgerdb \
  expectdigital/translate-agent-all-in-one:latest
```
Remove existing, pull latest and run
```bash
docker rm -f translate-all-in-one 2> /dev/null; docker pull expectdigital/translate-agent-all-in-one; docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID={Google project id} \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION={Google location} \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID={AWS access key id} \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY={AWS secret access key} \
  -e TRANSLATE_OTHER_AWS_REGION={AWS region} \
  -v path/to/google_account_key.json:/app/google_account_key.json \
  -v path/to/badger-dir:/data/badgerdb \
  expectdigital/translate-agent-all-in-one
```

### All-in-one image docker run arguments description
| Argument                                                              | Description                                            |
|-----------------------------------------------------------------------|--------------------------------------------------------|
| `-p 8080:8080`                                                        | Translate service port                                 |
| `-p 16686:16686`                                                      | Jaeger UI port                                         |
| `-e TRANSLATE_OTHER_GOOGLE_PROJECT_ID={Google project id}`            | Google project id                                      |
| `-e TRANSLATE_OTHER_GOOGLE_LOCATION={Google location}`                | Google  location                                       |
| `-e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID={AWS access key id}`            | AWS access key id                                      |
| `-e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY={AWS secret access key}     | AWS secret access key                                  |
| `-e TRANSLATE_OTHER_AWS_REGION={AWS region}`                          | AWS region                                             |
| `-v path/to/google_account_key.json:/app/google_account_key.json`     | Path to Google account key                             |
| `-v path/to/badger-dir:/data/badgerdb`                                | Path for BadgerDB db for data persistency *(Optional)* |
| `-v path/to/envoy.yaml:/app/envoy.yaml`                               | Path to custom envoy.yaml *(Optional)*                 |

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
