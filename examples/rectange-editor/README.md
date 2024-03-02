# Rectangle editor

This is a very simple sample application built on top of a «one-pass» IMGUI implementation, in which every frame every widget is called exactly once.

The sample application is a «rectangle editor». You can move rectangles by dragging them, resize them by dragging their edges and corners, and change their color using the provided RGB sliders. Deselect by clicking on the background. Obvious this is not a very interesting program in practice, but it offers a straightforward demonstration.

The application demonstrates various UI features to show that they can be easily implemented in IMGUI:

 * overlapping widgets ('hot_to_be' detects top-most widget)
 * dynamic widget presence (based on selecting rectangles)
 * dynamic widget sizing (rescale to 1/4 width of screen)
 * dynamic widget configuration ('show values' button)
 * custom widgets (the rectangle editor is effectively one)
 * widget animation (the fade-in/out when hovering over)

![image](https://github.com/anton2920/imgo/assets/30488779/fb804919-d18c-4056-b697-eace9c014c2f)

## Copyright

Pavlovskii Anton, 2024 (MIT). See [LICENSE](LICENSE) for more details.
