# s3glacier-go
A simple [S3 Glacier](http://aws.amazon.com/glacier) client written in Go. Most of the code is derived from Glacier's [sdk-for-go](https://docs.aws.amazon.com/sdk-for-go/api/) documentation where you can find more information.

## Why Yet Another Glacier Client?
There are already many clients available to interact with the Glacier APIs directly. For example, [Boto](https://github.com/boto/boto3) is a Glacier client written in Python, and there is also this [bash CLI tool](https://github.com/carlossg/glacier-cli). 

The short answer is why not? I'm learning golang at this moment, and I really enjoy developing something in go that is useful to me (and potentiallly to others as well). Also, since I'm running this tool in a Pi4 machine, Golang provides a few advantages to other languages I know, like Java and Python. It typically uses less memory (especially compared with JVM based languages), and out of the box it provides a nice package management tool that compiles all codes and dependencies into a standalone executable file so I don't need to install additional software to run it (you may be able to do that with 3rd party tools for Java and Python as well).

## Design Philosophy
As I mentioned earlier, I built this tool to fit my own use cases, and that dictates my decision making. I hope it'd be helpful to you as well, but not necessarily so. So please read the following before you use it.

### Command Line Based
One of the biggest drawback of interfacing with Glacier's APIs directly is that it doesn't provide immediate information on your archives - you may need to fire up a list-archive job and wait for the response to see what's in your vault. For this purpose, S3 provides an easier way to upload archives to S3 buckets first, then automatically move the data into Glacier with time based policies. You can find more information on [this article](https://www.msp360.com/resources/blog/compare-amazon-glacier-direct-upload-and-glacier-upload-through-amazon-s3/). However, I personally prefer to upload to Glacier directly without the need to deal with S3 buckets.

### Multipart Upload by Default
[Multipart Upload](https://docs.aws.amazon.com/amazonglacier/latest/dev/uploading-archive-mpu.html) is enabled by default. I back up all my personal files into a giant 7z archive (several 100s of GBs), and incrementally update it once in awhile. Multipart upload provides a list of advantages to upload bigger archives onto Glacier:

* If one part upload fails, it doesn't fail the entire upload, the failed part can be simply retried at a later time
* A single upload has the limit of 4GB, while you can upload an archive up to 40TB with multipart upload, according to [this article](https://docs.aws.amazon.com/amazonglacier/latest/dev/uploading-an-archive.html).
* For my home network, 1GB takes about 30-min to upload, so it's the right size for each part (the chunk size). This is the default chunk size. You can pick a chunk size that's right for you - for example, if your network is fast and stable, you can pick a larger chunk size up to 4GB. I believe each chunk upload is counted as 1 PUT request, and according to [this article](https://aws.amazon.com/s3/pricing/), you pay $0.03 for every 1000 requests. Thus unless each chunk is 1MB, the increased chunk count should incur negligible costs.

### State is Kept Locally with a MySQL
Like I mentioned earlier, Glacier doesn't provide an easy interface to interact with your archives directly. Thus I'm using a MySQL DB to persist states locally. This way, I can check the status of my uploads, retry failed ones, and when I want to retrieve my archives, I can specify the ArchiveId to download it directly.

## Prerequisite
* Set up your AWS account and an Admin user following this [Official Guide](https://docs.aws.amazon.com/amazonglacier/latest/dev/getting-started-before-you-begin.html#setup).
* Download the compiled binary for linux arm (good for Raspberry 4 Pi OS) or compile your own
* (Optional) Download the [Golang](https://golang.org/) binary that suits your platform.
* (Optional) Download this repository
* (Optional) In the root folder of the downloaded repository, run `env GOOS=<your_target_os> GOARCH=<your_target_arch> go build` (for a list of supported architecture, see [this page](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63)). This will generate the executable file `s3glaicer`.
* Install a MySQL server, and create a database.
* Use the included `create_tables.sql` to create necessary tables.

## Command line options
General form: `s3glacier <program_name> <options>`. You can see all available programs by running `s3glaicer`, or see all options under a program by running `s3glacier <program_name>`.

### s3glacier archive-upload
Upload your local archives to a Glacier's vault with multipart-upload.
```
 -v         The name of the vault to upload the archive to
 -f         The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file
 -u         The username of the MySQL database
 -p         The password of the MySQL database
 -db        The name of the database created
 -p         The IP address and port number of the database, defaults to `localhost:3306`
 -uploadId  The id of the upload (from the `uploads` table) to resume, if some of its parts had failed to be uploaded previously
 -s         The size of each chunk, defaults to 1GB (1024 * 1024 * 1024 bytes)
```

### s3glacier checksum-check
Check the checksum of a chunk of a file for debugging purposes.
```
 -f     The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file
 -o     The offset in bytes to read the file with, defaults to 0 (start of the file)
 -e     The expected checksum, defaults to empty
```

### s3glacier retrieve-archive
Still work in progress.
```
 -a     The id of the archive to retrieve
```

## Disclaimer
* Comments are welcome
* I'm not affiliated with Amazon
* Use this tool at your own risk
* S3 charges money on number of requests, disk usage, as well as retrieval, so make sure you understand their [pricing tiers](https://aws.amazon.com/s3/pricing/).

## License
```
  Copyright 2021 Xinxin Dai

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
```