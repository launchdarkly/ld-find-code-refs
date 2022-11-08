../ld-find-code-refs/ld-find-code-refs \
  --accessToken api-a0e12000-406f-4a9f-ba37-56b97b6b01d3 \
  --baseUri https://app.ld.catamorphic.com \
  --projKey default \
  --repoName accelerate \
  --repoType github \
  --dir ./ \
  --userAgent local \
  --dryRun \
  --debug

// 2022 api-5dfea03a-6917-4c32-b418-e1665378a751

// 2021 api-daaeeb15-f8b8-4439-81cc-7ec95ac80a38

//other tests api-a0e12000-406f-4a9f-ba37-56b97b6b01d3

act -j findCodeRefs --container-architecture linux/amd64 -s LD_ACCESS_TOKEN=api-a0e12000-406f-4a9f-ba37-56b97b6b01d3


./ld-find-code-refs \
  --accessToken api-a0e12000-406f-4a9f-ba37-56b97b6b01d3 \
  --baseUri https://app.ld.catamorphic.com \
  --projKey default \
  --repoName gonfalon \
  --repoType github \
  --dir ./ \
  --userAgent local \
  --debug

./../ld-find-code-refs/ld-find-code-refs \
  --accessToken api-a0e12000-406f-4a9f-ba37-56b97b6b01d3 \
  --baseUri https://app.ld.catamorphic.com \
  --projKey default \
  --repoName gonfalon \
  --repoType github \
  --dir ./ \
  --userAgent local \
  --debug
