# flac-to-aac-library-converter

This Go program is a tool I wrote to convert a large library of FLAC music files to AAC format (while not re-encoding, just copying any MP3s), so I could use them on my iPod classic. I have a personal library of around 10,000 tracks, and this program provides an efficient way to convert these files using a worker pool approach.

In addition to FLAC files, this tool also supports copying MP3 files without conversion, maintaining the directory structure from the source directory to the destination directory.

## Prerequisites
The FLAC to AAC conversion uses qaac64.exe, so make sure you have it installed and available in your system path. You can download it from [here](https://github.com/nu774/qaac/releases) and follow the instructions to install it.

## Usage

```bash
flac_to_aac_library_converter.exe --src <source_directory> --dest <destination_directory> --workers <num_workers>
```

- source_directory: The directory containing the FLAC and MP3 files you want to convert or copy.
- destination_directory: The directory where the converted or copied files will be saved.
- num_workers (optional): The number of workers to process the files concurrently. Default is 5.
