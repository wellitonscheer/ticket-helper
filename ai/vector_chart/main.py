import matplotlib.pyplot as plt
import numpy as np

word_embeddings = {
    "rei": (4, 2),
    "rainha": (3, 1),
    "cachorro": (1, 5),
    "gato": (2, 4),
    "castelo": (4, 1),
}

king_coords = word_embeddings["rei"]
cat_coords = word_embeddings["gato"]

king_mid = (king_coords[0] / 2, king_coords[1] / 2)
cat_mid = (cat_coords[0] / 2, cat_coords[1] / 2)

control_point = ((king_mid[0] + cat_mid[0]) / 2 + 0.5, (king_mid[1] + cat_mid[1]) / 2 + 0.5)

t = np.linspace(0, 1, 100)
curve_x = (1 - t) ** 2 * king_mid[0] + 2 * (1 - t) * t * control_point[0] + t ** 2 * cat_mid[0]
curve_y = (1 - t) ** 2 * king_mid[1] + 2 * (1 - t) * t * control_point[1] + t ** 2 * cat_mid[1]


plt.figure(figsize=(5, 5))
for word, (x, y) in word_embeddings.items():
    plt.scatter(x, y, color="blue")
    plt.text(x + 0.1, y + 0.1, word, fontsize=12)
    plt.plot([0, x], [0, y], color="green")

# circle = plt.Circle((4, 1), 1.1, color='red', fill=False, linewidth=1)
# plt.gca().add_patch(circle)
plt.plot(curve_x, curve_y, color="red", linewidth=2)

plt.xlim(0, 6)
plt.ylim(0, 6)

plt.xlabel("X")
plt.ylabel("Y")

plt.grid(True, linestyle="--", alpha=0.6)

plt.title("2D Word Embeddings")
plt.show()
