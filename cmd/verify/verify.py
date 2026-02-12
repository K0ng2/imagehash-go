import imagehash
from PIL import Image

img = Image.open('image.png')
print(f'ahash: {imagehash.average_hash(img)}')
print(f'phash: {imagehash.phash(img)}')
print(f'dhash: {imagehash.dhash(img)}')
print(f'dhash_v: {imagehash.dhash_vertical(img)}')
