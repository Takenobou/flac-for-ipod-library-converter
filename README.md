# flac-for-ipod-library-converter

This Go program is a tool I wrote to convert a large library of FLAC music files to AAC and Opus format, so I could use
them on my iPod classic. I have a personal library of around 10,000 tracks, and this program provides an efficient way
to convert these files using a worker pool approach.

In addition to FLAC files, this tool also supports copying MP3 files without conversion, maintaining the directory 
structure from the source directory to the destination directory.

Since Opus does not support embedded album art, this tool will take the album art from a file in the album directory 
with the name "cover.png" and convert it to a baseline 320x320 jpg as this works best with Rockbox (which is required 
to play Opus files on the iPod classic).

From my testing, I find that 160kbps is a good bitrate for Opus files, and 192kbps is a good bitrate for AAC files so
these are the defaults. You can override these defaults by specifying the bitrate parameter. 

Opus produces a much smaller library than AAC (at 10,000 tracks AAC takes 63GB and Opus takes 53GB) which matters when 
you are transferring over usb 2.0 speeds (haha), but isnt supported by the stock firmware.


## Prerequisites
The FLAC to AAC conversion uses qaac64.exe, so make sure you have it installed and available in your system path.
You can download it from [here](https://github.com/nu774/qaac/releases) and follow the instructions to install it.

The FLAC to Opus conversion uses opusenc.exe, so make sure you have it installed and available in your system path.
You can download it from [here](https://opus-codec.org/downloads/) and follow the instructions to install it.

## Usage

```bash
flac_for_ipod_library_converter.exe --src <source_directory> --dest <destination_directory> --workers <num_workers> --codec <codec> --bitrate <bitrate>
```

- source_directory: The directory containing the FLAC and MP3 files you want to convert or copy.
- destination_directory: The directory where the converted or copied files will be saved.
- num_workers (optional): The number of workers to process the files concurrently. Default is 5.
- codec (optional): The codec to use for conversion. Valid values are "aac" and "opus". Default is "aac".
- bitrate (optional): The bitrate to use for conversion. Valid values are "128", "192", "256", and "320". 
Default is "192" for aac and "160"  for opus.