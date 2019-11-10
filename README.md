# cdn77-refresh

## Getting started
CDN77 is a content delivery network run by DataCamp Limited (UK). [CDN77 API](https://client.cdn77.com/support/api) allows you to run a wide range of commands and tasks from an external script or server on CDN77.

The basic idea of this tool is to make CDN77 aware that there is new content available, so it removes ("purges") all data for a specific site / CDN resource from CDN77, tries to fetch sitemap.txt, and make CDN77 preload the content of the URLs found.


## Building and dependencies
[UPX](https://upx.github.io) and Unix make are required to build.

Please run ```make``` to build.

## Running
```
gocdn77-refresh --login=LOGIN --token=TOKEN --site=SITE --sitemap=<sitemap.xml> --purge-all --verbose
```
* --login : Your login (email) to CDN77 control panel
* --token : Your API Token, needs to be generated in the profile section on client.CDN77.com
* --site : Your website aka 'CDN Resource' in CDN77
* --purge-all : remove (purge) existing HTTP content on CDN77
* --verbose : additional output

Typical run would be:
```
2019-06-07T08:00:34+02:00 | INFO  | Reading resource list from CDN77 ... ok
2019-06-07T08:00:34+02:00 | INFO  | Searching for (...) ... ok (resource id #...)
2019-06-07T08:00:35+02:00 | INFO  | Starting 'purge-all' ... ok
2019-06-07T08:00:35+02:00 | INFO  | Reading sitemap.xml ... ok
2019-06-07T08:00:35+02:00 | INFO  | Prefetching (...) ... ok
(...)
```

## Reference
* [CDN77 API documentation](https://client.cdn77.com/support/api)

## License
Released under the [MIT  License](https://en.wikipedia.org/wiki/MIT_License)