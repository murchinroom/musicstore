# musicstore

Your "Vinyl Record Storage Shelf".

> musicstore = audiofilestore + crud(metadata) + murecom

## Usage

### Run the server

```
vim config.yaml # edit the config file
go run .        # -h for help
```

### Get tracks

Get all tracks:

```sh
curl localhost:8080/tracks  # | json_pp
```

Get a specific track by ID:

```sh
curl localhost:8080/tracks/1
```

(Endpoint `/tracks` supports other RESFful CRUD operations.)

### Post new tracks

Upload a file:

```sh
curl -X POST -F 'File=@song.mp3' localhost:8080/example-audio/new
```

Or let server download the file from a URL:

```sh
curl -X POST -F 'AudioFileURL=https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3' localhost:8080/example-audio/new
```

### Emotion based music recommendation

Get a recommendation based on your current emotion:

```sh
curl 'localhost:8080/murecom?Valence=0.5&Arousal=0.5'
```
