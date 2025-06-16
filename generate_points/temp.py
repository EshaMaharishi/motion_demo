import numpy as np
from matplotlib.textpath import TextPath

# Adjust font size and spacing as needed
size = 45
spacing = 30

text = input("Enter the text to convert into 3D points: ")

x_offset = 0.0

with open("points3d.txt", "w") as f:
    f.write("x, y, z  # letter\n")

    for letter in text:
        print("Writing points for letter", letter)

        tp = TextPath((0, 0), letter, size=size)
        points_2d = tp.vertices

        # Make points_2d writeable
        points_2d = np.array(tp.vertices, copy=True)

        # Offset the x-coordinates
        points_2d[:, 0] += x_offset

        # Project to 3D
        points_3d = np.hstack([points_2d, np.zeros((points_2d.shape[0], 1))])

        # Write visual divider and label
        f.write(f"# --- Letter: {letter} ---\n")

        # Write points for the current letter
        for pt in points_3d:
            f.write(f"{pt[0]:.4f}, {pt[1]:.4f}, {pt[2]:.4f}\n")

        # Update offset for next letter
        x_offset += spacing

print("3D points written to points3d.txt")
