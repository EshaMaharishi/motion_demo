import numpy as np
from matplotlib.textpath import TextPath

# Create the text path
text = input("Enter the text to convert into 3D points: ")
tp = TextPath((0, 0), text, size=1)
points_2d = tp.vertices

# Project to 3D (z = 0)
points_3d = np.hstack([points_2d, np.zeros((points_2d.shape[0], 1))])

# Normalize for color mapping
num_points = points_3d.shape[0]
colors = np.linspace(0, 1, num_points)

# Write points to file
with open("points3d.txt", "w") as f:
    f.write("x, y, z\n")
    for pt in points_3d:
        f.write(f"{pt[0]:.4f}, {pt[1]:.4f}, {pt[2]:.4f}\n")

print("3D points written to points3d.txt")
