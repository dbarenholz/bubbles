# Bubbles

A collection of my [bubbletea](https://github.com/charmbracelet/bubbletea) [bubbles](https://github.com/charmbracelet/bubbles).
These have been created due to my need for reusable components in TUIs.

## TraversableFilepicker

The **TraversableFilepicker** bubble is effectively the [charm filepicker bubble](https://github.com/charmbracelet/bubbles/tree/main/filepicker), but with a different API that makes it impossible to get "stuck" in a directory (e.g., you can no longer set `CurrentDirectory` to `.`, preventing `Back` from working).
Essentially all functionality of the charm filepicker is preserved, but we have following additions:

1. `Home` and `End` keys are now mapped by default to `GoToTop` and `GoToBottom` actions, respectively.
2. Go to your user home by pressing `~`.
3. Allow `LoopEntries` to loop the selection when going out of bounds; if you keep holding the down arrow (or `j` for vim enthusiasts), the selection will continuously loop through the entries instead of stopping at the end (if `LoopEntries = true`). Defaults to `false`.

See the [example](./examples/TraversableFilepicker/README.md) for details and a pretty gif.

## License

This repository and its bubbles are [MIT](./LICENSE.md) licensed.

- [Bubbletea](https://github.com/charmbracelet/bubbletea) is MIT licensed.
- [Bubbles](https://github.com/charmbracelet/bubbles) is MIT licensed.
