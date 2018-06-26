## Notes for translator authors

### Variable params substitution

Some xlators require non-static information such as brick path or volume id to
be populated by glusterd2 in xlator's options in volfile. Xlators can set a
placeholder/varstring in the `.default_value` field of its options table.
Glusterd2 will replace this placeholder/varstring with actual values during
volfile generation.

Here's a list of supported varstrings:

```
volume.id
volume.name
volume.type
volume.redundancy
volume.transport
volume.auth.username
volume.auth.password
```
```
brick.id
brick.hostname
brick.peerid
brick.path
brick.volumename
brick.volumeid
```

Example code from `xlators/features/index/src/index.c`:

```c
struct volume_options options[] = {
        { .key  = {"index-base" },
          .type = GF_OPTION_TYPE_PATH,
          .description = "path where the index files need to be stored",
          .default_value = "{{ brick.path }}/.glusterfs/indices"
        },
...
}
```

The varstring `{{ brick.path }}` will be substituted with actual path of the
brick during volfile generation. This is how it'll end up in the generated
brick volfile:

```
volume test-index
    type features/index
    option index-base /export/brick2/data/.glusterfs/indices
    option xattrop-pending-watchlist trusted.afr.test
    subvolumes test-barrier
end-volume
```
