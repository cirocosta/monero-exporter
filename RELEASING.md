# releasing

1. build the container images

```
make images
```

2. create a final commit with the images checked out

```
git add --all .
git commit # bla bla
```

3. create a new signed tag

```
git tag -s $version
```

4. build the release binaries

```
goreleaser release --rm-dist
```

5. sign the checksums

```
export GPG_TTY=$(tty)
gpg --clearsign ./dist/checksums.txt
```
