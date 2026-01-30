from PIL import Image

# Load original image
img_path = "/Users/fernando_idwell/Projects/Skillsphere/skillsphere/skillsphere-pwa/assets/static/assets/static/lol.png"
img = Image.open(img_path).convert("RGBA")

# Resize and save
sizes = [192, 512]
out_paths = []

for size in sizes:
    resized = img.resize((size, size), Image.LANCZOS)
    out_path = f"/Users/fernando_idwell/Projects/Skillsphere/skillsphere/skillsphere-pwa/assets/static/assets/static/icon-{size}.png"
    resized.save(out_path, format="PNG")
    out_paths.append(out_path)

out_paths
