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

## SearchableFilepicker

The **SearchableFilepicker** bubble is a wrapper around the TraversableFilepicker that adds a search input above it, allowing you to filter the displayed files and directories based on the search query.
Filtering is done real-time, but is _not_ fuzzy; it is a dumb case-insensitive substring match.
As addition, the path of the current directory is displayed above the filepicker.
This to illustrate how one can embed the filepicker in a larger UI and still have it work as expected.

See the [example](./examples/SearchableFilepicker/README.md) for details and a pretty gif.

## Button

The **Button** bubble renders stylable and composable text "buttons".
A simple style is to use `[ text ]` as button, call it "bracket style".
Buttons have states for enabled, disabled, and being focused/selected.
Buttons do indeed behave like buttons, and can be pressed by hitting the `enter` when focused, which fires a `ButtonPressed` message that can be handled by parent UIs.

See the [example](./examples/Button/README.md) for details and a pretty gif.

## ButtonGroup

The **ButtonGroup** bubble is a container for buttons, showcasing their composability.
You either create a horizontal or vertical group, or you specify the number of rows and columns for a grid layout.
Spacing is customizable, but has sensible (insofar possible) defaults.
Focus between buttons in a group is done using arrow keys, or using `tab` and `shift+tab` if you so prefer.
The `ButtonPressed` message from buttons in the group are given an extra index for a group, so it's easier to identify which button was pressed.

See the [example](./examples/ButtonGroup/README.md) for details and a pretty gif.

## License

This repository and its bubbles are [MIT](./LICENSE.md) licensed.

- [Bubbletea](https://github.com/charmbracelet/bubbletea) is MIT licensed.
- [Bubbles](https://github.com/charmbracelet/bubbles) is MIT licensed.
