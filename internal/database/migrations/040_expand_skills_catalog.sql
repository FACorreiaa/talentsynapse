-- +goose Up
-- Expand skill categories and skills to be more generalist
-- This migration adds comprehensive categories covering tech, sports, cooking, construction, arts, and more

-- 1. Insert Additional Categories
INSERT INTO skill_categories (name, description) VALUES
    -- Technology & Professional
    ('Technology', 'Software, hardware, and digital skills'),
    ('Data & Analytics', 'Data science, analytics, and business intelligence'),
    ('Design', 'Graphic design, UX/UI, and visual arts'),
    ('Business', 'Business management, entrepreneurship, and leadership'),
    ('Marketing', 'Digital marketing, content creation, and branding'),
    ('Finance', 'Personal finance, investing, and accounting'),

    -- Trades & Construction
    ('Construction', 'Building, renovation, and structural work'),
    ('Electrical', 'Electrical installation and repair'),
    ('Plumbing', 'Plumbing installation and maintenance'),
    ('Automotive', 'Vehicle repair and maintenance'),
    ('Woodworking', 'Carpentry and furniture making'),
    ('Metalworking', 'Welding, fabrication, and metalcraft'),

    -- Sports & Fitness
    ('Team Sports', 'Team-based athletic activities'),
    ('Individual Sports', 'Solo athletic pursuits'),
    ('Fitness', 'Exercise, strength training, and wellness'),
    ('Martial Arts', 'Combat sports and self-defense'),
    ('Water Sports', 'Swimming, surfing, and aquatic activities'),
    ('Outdoor Activities', 'Hiking, camping, and adventure sports'),

    -- Culinary
    ('Cooking', 'Culinary arts and food preparation'),
    ('Baking', 'Pastry, bread, and dessert making'),
    ('Beverages', 'Coffee, cocktails, and drink preparation'),
    ('Nutrition', 'Diet planning and healthy eating'),

    -- Creative Arts
    ('Visual Arts', 'Painting, drawing, and sculpture'),
    ('Photography', 'Photography and photo editing'),
    ('Film & Video', 'Video production and editing'),
    ('Writing', 'Creative writing, copywriting, and journalism'),
    ('Crafts', 'Handmade crafts and DIY projects'),
    ('Fashion', 'Fashion design and styling'),

    -- Performing Arts
    ('Dance', 'Various dance styles and choreography'),
    ('Theater', 'Acting and theatrical performance'),
    ('Voice', 'Singing and vocal training'),

    -- Education & Personal Development
    ('Teaching', 'Education and tutoring skills'),
    ('Public Speaking', 'Presentation and communication'),
    ('Personal Development', 'Life skills and self-improvement'),
    ('Parenting', 'Childcare and family skills'),

    -- Health & Wellness
    ('Health', 'Health and medical knowledge'),
    ('Mental Wellness', 'Meditation, mindfulness, and stress management'),
    ('Beauty', 'Skincare, makeup, and personal grooming'),

    -- Home & Garden
    ('Gardening', 'Plant care and landscaping'),
    ('Home Improvement', 'DIY home projects and maintenance'),
    ('Interior Design', 'Home decoration and space planning'),
    ('Cleaning', 'Professional cleaning and organization'),

    -- Science & Engineering
    ('Science', 'Scientific knowledge and research'),
    ('Engineering', 'Engineering principles and applications'),
    ('Environment', 'Sustainability and environmental practices'),

    -- Gaming & Entertainment
    ('Gaming', 'Video games and esports'),
    ('Board Games', 'Tabletop and strategy games'),

    -- Animals & Nature
    ('Pet Care', 'Animal care and training'),
    ('Agriculture', 'Farming and animal husbandry')
ON CONFLICT (name) DO NOTHING;

-- 2. Insert Skills for each category

-- Technology Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Technology')
INSERT INTO skills (name, category_id, description) VALUES
    ('Software Development', (SELECT id FROM cat), 'Building software applications and systems'),
    ('Web Development', (SELECT id FROM cat), 'Creating websites and web applications'),
    ('Mobile App Development', (SELECT id FROM cat), 'Building iOS and Android applications'),
    ('Cloud Computing', (SELECT id FROM cat), 'AWS, Azure, GCP, and cloud infrastructure'),
    ('DevOps', (SELECT id FROM cat), 'CI/CD, automation, and infrastructure management'),
    ('Cybersecurity', (SELECT id FROM cat), 'Security practices and threat prevention'),
    ('Database Administration', (SELECT id FROM cat), 'Managing and optimizing databases'),
    ('System Administration', (SELECT id FROM cat), 'Managing servers and IT infrastructure'),
    ('Network Engineering', (SELECT id FROM cat), 'Designing and maintaining computer networks'),
    ('IT Support', (SELECT id FROM cat), 'Technical support and troubleshooting'),
    ('Blockchain', (SELECT id FROM cat), 'Blockchain technology and smart contracts'),
    ('AI & Machine Learning', (SELECT id FROM cat), 'Artificial intelligence and ML models'),
    ('Robotics', (SELECT id FROM cat), 'Building and programming robots'),
    ('IoT Development', (SELECT id FROM cat), 'Internet of Things devices and systems'),
    ('Game Development', (SELECT id FROM cat), 'Creating video games'),
    ('3D Printing', (SELECT id FROM cat), '3D printing and additive manufacturing')
ON CONFLICT (category_id, name) DO NOTHING;

-- Data & Analytics Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Data & Analytics')
INSERT INTO skills (name, category_id, description) VALUES
    ('Data Science', (SELECT id FROM cat), 'Extracting insights from data'),
    ('Data Analysis', (SELECT id FROM cat), 'Analyzing and interpreting data'),
    ('Business Intelligence', (SELECT id FROM cat), 'BI tools and reporting'),
    ('Data Visualization', (SELECT id FROM cat), 'Creating charts and dashboards'),
    ('Statistical Analysis', (SELECT id FROM cat), 'Statistical methods and modeling'),
    ('SQL & Databases', (SELECT id FROM cat), 'Database querying and management'),
    ('Excel & Spreadsheets', (SELECT id FROM cat), 'Advanced spreadsheet skills'),
    ('Power BI', (SELECT id FROM cat), 'Microsoft Power BI dashboards'),
    ('Tableau', (SELECT id FROM cat), 'Tableau visualization platform')
ON CONFLICT (category_id, name) DO NOTHING;

-- Design Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Design')
INSERT INTO skills (name, category_id, description) VALUES
    ('Graphic Design', (SELECT id FROM cat), 'Visual communication and branding'),
    ('UI Design', (SELECT id FROM cat), 'User interface design'),
    ('UX Design', (SELECT id FROM cat), 'User experience research and design'),
    ('Logo Design', (SELECT id FROM cat), 'Creating brand logos and marks'),
    ('Illustration', (SELECT id FROM cat), 'Digital and traditional illustration'),
    ('Motion Graphics', (SELECT id FROM cat), 'Animated graphics and visual effects'),
    ('3D Modeling', (SELECT id FROM cat), 'Creating 3D models and renders'),
    ('Product Design', (SELECT id FROM cat), 'Designing physical products'),
    ('Adobe Photoshop', (SELECT id FROM cat), 'Photo editing and manipulation'),
    ('Adobe Illustrator', (SELECT id FROM cat), 'Vector graphics creation'),
    ('Figma', (SELECT id FROM cat), 'Collaborative design tool'),
    ('Canva', (SELECT id FROM cat), 'Simplified graphic design')
ON CONFLICT (category_id, name) DO NOTHING;

-- Business Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Business')
INSERT INTO skills (name, category_id, description) VALUES
    ('Project Management', (SELECT id FROM cat), 'Planning and executing projects'),
    ('Product Management', (SELECT id FROM cat), 'Managing product development'),
    ('Entrepreneurship', (SELECT id FROM cat), 'Starting and running businesses'),
    ('Leadership', (SELECT id FROM cat), 'Team leadership and management'),
    ('Strategic Planning', (SELECT id FROM cat), 'Business strategy development'),
    ('Operations Management', (SELECT id FROM cat), 'Managing business operations'),
    ('Sales', (SELECT id FROM cat), 'Selling products and services'),
    ('Customer Service', (SELECT id FROM cat), 'Supporting and helping customers'),
    ('Human Resources', (SELECT id FROM cat), 'HR management and recruiting'),
    ('Consulting', (SELECT id FROM cat), 'Business consulting and advisory'),
    ('Negotiation', (SELECT id FROM cat), 'Negotiation strategies and tactics'),
    ('Agile & Scrum', (SELECT id FROM cat), 'Agile methodologies and frameworks')
ON CONFLICT (category_id, name) DO NOTHING;

-- Marketing Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Marketing')
INSERT INTO skills (name, category_id, description) VALUES
    ('Digital Marketing', (SELECT id FROM cat), 'Online marketing strategies'),
    ('Social Media Marketing', (SELECT id FROM cat), 'Marketing on social platforms'),
    ('Content Marketing', (SELECT id FROM cat), 'Creating valuable content'),
    ('SEO', (SELECT id FROM cat), 'Search engine optimization'),
    ('Email Marketing', (SELECT id FROM cat), 'Email campaigns and automation'),
    ('PPC Advertising', (SELECT id FROM cat), 'Pay-per-click advertising'),
    ('Influencer Marketing', (SELECT id FROM cat), 'Working with influencers'),
    ('Brand Strategy', (SELECT id FROM cat), 'Building and managing brands'),
    ('Copywriting', (SELECT id FROM cat), 'Writing persuasive content'),
    ('Analytics & Metrics', (SELECT id FROM cat), 'Marketing analytics and reporting'),
    ('Video Marketing', (SELECT id FROM cat), 'Creating marketing videos'),
    ('Podcast Production', (SELECT id FROM cat), 'Creating and producing podcasts')
ON CONFLICT (category_id, name) DO NOTHING;

-- Finance Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Finance')
INSERT INTO skills (name, category_id, description) VALUES
    ('Personal Finance', (SELECT id FROM cat), 'Managing personal money'),
    ('Investing', (SELECT id FROM cat), 'Investment strategies and analysis'),
    ('Accounting', (SELECT id FROM cat), 'Financial accounting and bookkeeping'),
    ('Budgeting', (SELECT id FROM cat), 'Creating and managing budgets'),
    ('Tax Preparation', (SELECT id FROM cat), 'Tax filing and planning'),
    ('Financial Planning', (SELECT id FROM cat), 'Long-term financial strategy'),
    ('Cryptocurrency', (SELECT id FROM cat), 'Crypto trading and investing'),
    ('Stock Trading', (SELECT id FROM cat), 'Trading stocks and securities'),
    ('Real Estate Investing', (SELECT id FROM cat), 'Property investment strategies')
ON CONFLICT (category_id, name) DO NOTHING;

-- Construction Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Construction')
INSERT INTO skills (name, category_id, description) VALUES
    ('General Construction', (SELECT id FROM cat), 'Building and construction basics'),
    ('Framing', (SELECT id FROM cat), 'Wood and metal framing'),
    ('Roofing', (SELECT id FROM cat), 'Roof installation and repair'),
    ('Drywall Installation', (SELECT id FROM cat), 'Installing and finishing drywall'),
    ('Masonry', (SELECT id FROM cat), 'Brick and stone work'),
    ('Concrete Work', (SELECT id FROM cat), 'Pouring and finishing concrete'),
    ('Flooring Installation', (SELECT id FROM cat), 'Installing various flooring types'),
    ('Tiling', (SELECT id FROM cat), 'Tile installation and grouting'),
    ('Insulation', (SELECT id FROM cat), 'Installing building insulation'),
    ('Demolition', (SELECT id FROM cat), 'Safe demolition practices'),
    ('Blueprint Reading', (SELECT id FROM cat), 'Reading construction plans')
ON CONFLICT (category_id, name) DO NOTHING;

-- Electrical Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Electrical')
INSERT INTO skills (name, category_id, description) VALUES
    ('Residential Electrical', (SELECT id FROM cat), 'Home electrical systems'),
    ('Commercial Electrical', (SELECT id FROM cat), 'Commercial electrical work'),
    ('Electrical Troubleshooting', (SELECT id FROM cat), 'Diagnosing electrical problems'),
    ('Lighting Installation', (SELECT id FROM cat), 'Installing light fixtures'),
    ('Panel Upgrades', (SELECT id FROM cat), 'Upgrading electrical panels'),
    ('Smart Home Installation', (SELECT id FROM cat), 'Installing smart home systems'),
    ('Solar Panel Installation', (SELECT id FROM cat), 'Installing solar systems'),
    ('EV Charger Installation', (SELECT id FROM cat), 'Installing electric vehicle chargers')
ON CONFLICT (category_id, name) DO NOTHING;

-- Plumbing Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Plumbing')
INSERT INTO skills (name, category_id, description) VALUES
    ('Residential Plumbing', (SELECT id FROM cat), 'Home plumbing systems'),
    ('Pipe Fitting', (SELECT id FROM cat), 'Installing and repairing pipes'),
    ('Drain Cleaning', (SELECT id FROM cat), 'Clearing blocked drains'),
    ('Water Heater Installation', (SELECT id FROM cat), 'Installing water heaters'),
    ('Fixture Installation', (SELECT id FROM cat), 'Installing sinks, toilets, etc.'),
    ('Leak Detection', (SELECT id FROM cat), 'Finding and fixing leaks'),
    ('Septic Systems', (SELECT id FROM cat), 'Septic system maintenance')
ON CONFLICT (category_id, name) DO NOTHING;

-- Automotive Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Automotive')
INSERT INTO skills (name, category_id, description) VALUES
    ('Auto Mechanics', (SELECT id FROM cat), 'General vehicle repair'),
    ('Engine Repair', (SELECT id FROM cat), 'Engine diagnostics and repair'),
    ('Brake Service', (SELECT id FROM cat), 'Brake system maintenance'),
    ('Oil Change', (SELECT id FROM cat), 'Oil and fluid changes'),
    ('Tire Service', (SELECT id FROM cat), 'Tire rotation and replacement'),
    ('Auto Electrical', (SELECT id FROM cat), 'Vehicle electrical systems'),
    ('Auto Body Repair', (SELECT id FROM cat), 'Body work and painting'),
    ('Auto Detailing', (SELECT id FROM cat), 'Vehicle cleaning and detailing'),
    ('Motorcycle Repair', (SELECT id FROM cat), 'Motorcycle maintenance'),
    ('Classic Car Restoration', (SELECT id FROM cat), 'Restoring vintage vehicles')
ON CONFLICT (category_id, name) DO NOTHING;

-- Woodworking Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Woodworking')
INSERT INTO skills (name, category_id, description) VALUES
    ('Carpentry', (SELECT id FROM cat), 'Building with wood'),
    ('Cabinet Making', (SELECT id FROM cat), 'Building cabinets and storage'),
    ('Furniture Making', (SELECT id FROM cat), 'Crafting furniture pieces'),
    ('Wood Carving', (SELECT id FROM cat), 'Decorative wood carving'),
    ('Wood Finishing', (SELECT id FROM cat), 'Staining and finishing wood'),
    ('Joinery', (SELECT id FROM cat), 'Wood joining techniques'),
    ('Wood Turning', (SELECT id FROM cat), 'Lathe work and turning'),
    ('Deck Building', (SELECT id FROM cat), 'Building outdoor decks')
ON CONFLICT (category_id, name) DO NOTHING;

-- Metalworking Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Metalworking')
INSERT INTO skills (name, category_id, description) VALUES
    ('Welding', (SELECT id FROM cat), 'Joining metals by welding'),
    ('Metal Fabrication', (SELECT id FROM cat), 'Creating metal structures'),
    ('Blacksmithing', (SELECT id FROM cat), 'Traditional metalwork'),
    ('Sheet Metal Work', (SELECT id FROM cat), 'Working with sheet metal'),
    ('CNC Machining', (SELECT id FROM cat), 'Computer-controlled machining'),
    ('Metal Casting', (SELECT id FROM cat), 'Casting metal objects'),
    ('Jewelry Making', (SELECT id FROM cat), 'Creating metal jewelry')
ON CONFLICT (category_id, name) DO NOTHING;

-- Team Sports Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Team Sports')
INSERT INTO skills (name, category_id, description) VALUES
    ('Soccer', (SELECT id FROM cat), 'Football/soccer skills and strategy'),
    ('Basketball', (SELECT id FROM cat), 'Basketball techniques and plays'),
    ('Baseball', (SELECT id FROM cat), 'Baseball and softball skills'),
    ('American Football', (SELECT id FROM cat), 'American football techniques'),
    ('Volleyball', (SELECT id FROM cat), 'Volleyball skills and tactics'),
    ('Hockey', (SELECT id FROM cat), 'Ice and field hockey'),
    ('Rugby', (SELECT id FROM cat), 'Rugby techniques and rules'),
    ('Cricket', (SELECT id FROM cat), 'Cricket batting and bowling'),
    ('Handball', (SELECT id FROM cat), 'Team handball skills'),
    ('Water Polo', (SELECT id FROM cat), 'Water polo techniques'),
    ('Lacrosse', (SELECT id FROM cat), 'Lacrosse skills and strategy')
ON CONFLICT (category_id, name) DO NOTHING;

-- Individual Sports Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Individual Sports')
INSERT INTO skills (name, category_id, description) VALUES
    ('Tennis', (SELECT id FROM cat), 'Tennis strokes and strategy'),
    ('Golf', (SELECT id FROM cat), 'Golf swing and course management'),
    ('Running', (SELECT id FROM cat), 'Distance and sprint running'),
    ('Cycling', (SELECT id FROM cat), 'Road and mountain biking'),
    ('Swimming', (SELECT id FROM cat), 'Swimming strokes and techniques'),
    ('Gymnastics', (SELECT id FROM cat), 'Gymnastic movements and routines'),
    ('Track & Field', (SELECT id FROM cat), 'Athletic events and training'),
    ('Badminton', (SELECT id FROM cat), 'Badminton techniques'),
    ('Table Tennis', (SELECT id FROM cat), 'Ping pong skills'),
    ('Squash', (SELECT id FROM cat), 'Squash techniques and strategy'),
    ('Bowling', (SELECT id FROM cat), 'Bowling technique and strategy'),
    ('Archery', (SELECT id FROM cat), 'Bow and arrow skills'),
    ('Fencing', (SELECT id FROM cat), 'Sword fighting sport'),
    ('Wrestling', (SELECT id FROM cat), 'Wrestling techniques'),
    ('Boxing', (SELECT id FROM cat), 'Boxing fundamentals'),
    ('Triathlon Training', (SELECT id FROM cat), 'Multi-sport endurance training')
ON CONFLICT (category_id, name) DO NOTHING;

-- Fitness Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Fitness')
INSERT INTO skills (name, category_id, description) VALUES
    ('Personal Training', (SELECT id FROM cat), 'Fitness coaching and programming'),
    ('Weight Training', (SELECT id FROM cat), 'Strength training techniques'),
    ('CrossFit', (SELECT id FROM cat), 'CrossFit workouts and movements'),
    ('Yoga', (SELECT id FROM cat), 'Yoga poses and sequences'),
    ('Pilates', (SELECT id FROM cat), 'Pilates exercises and instruction'),
    ('HIIT Training', (SELECT id FROM cat), 'High-intensity interval training'),
    ('Calisthenics', (SELECT id FROM cat), 'Bodyweight exercises'),
    ('Stretching & Flexibility', (SELECT id FROM cat), 'Flexibility training'),
    ('Aerobics', (SELECT id FROM cat), 'Aerobic exercise instruction'),
    ('Spinning', (SELECT id FROM cat), 'Indoor cycling classes'),
    ('Kettlebell Training', (SELECT id FROM cat), 'Kettlebell workouts'),
    ('Olympic Lifting', (SELECT id FROM cat), 'Olympic weightlifting'),
    ('Powerlifting', (SELECT id FROM cat), 'Powerlifting techniques'),
    ('Bodybuilding', (SELECT id FROM cat), 'Muscle building and posing')
ON CONFLICT (category_id, name) DO NOTHING;

-- Martial Arts Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Martial Arts')
INSERT INTO skills (name, category_id, description) VALUES
    ('Karate', (SELECT id FROM cat), 'Japanese striking art'),
    ('Judo', (SELECT id FROM cat), 'Japanese throwing art'),
    ('Taekwondo', (SELECT id FROM cat), 'Korean martial art'),
    ('Brazilian Jiu-Jitsu', (SELECT id FROM cat), 'Ground fighting and grappling'),
    ('Muay Thai', (SELECT id FROM cat), 'Thai kickboxing'),
    ('Kickboxing', (SELECT id FROM cat), 'Stand-up combat sport'),
    ('MMA', (SELECT id FROM cat), 'Mixed martial arts'),
    ('Kung Fu', (SELECT id FROM cat), 'Chinese martial arts'),
    ('Aikido', (SELECT id FROM cat), 'Japanese defensive art'),
    ('Krav Maga', (SELECT id FROM cat), 'Israeli self-defense system'),
    ('Capoeira', (SELECT id FROM cat), 'Brazilian martial art with dance'),
    ('Self-Defense', (SELECT id FROM cat), 'Practical self-defense techniques')
ON CONFLICT (category_id, name) DO NOTHING;

-- Water Sports Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Water Sports')
INSERT INTO skills (name, category_id, description) VALUES
    ('Surfing', (SELECT id FROM cat), 'Wave riding techniques'),
    ('Scuba Diving', (SELECT id FROM cat), 'Underwater diving'),
    ('Snorkeling', (SELECT id FROM cat), 'Surface snorkeling'),
    ('Kayaking', (SELECT id FROM cat), 'Kayak paddling and navigation'),
    ('Canoeing', (SELECT id FROM cat), 'Canoe techniques'),
    ('Stand-Up Paddleboarding', (SELECT id FROM cat), 'SUP techniques'),
    ('Sailing', (SELECT id FROM cat), 'Sailing boats and yachts'),
    ('Wakeboarding', (SELECT id FROM cat), 'Wakeboarding tricks'),
    ('Waterskiing', (SELECT id FROM cat), 'Water ski techniques'),
    ('Kitesurfing', (SELECT id FROM cat), 'Kite-powered surfing'),
    ('Windsurfing', (SELECT id FROM cat), 'Wind-powered board sailing'),
    ('Rowing', (SELECT id FROM cat), 'Rowing and sculling')
ON CONFLICT (category_id, name) DO NOTHING;

-- Outdoor Activities Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Outdoor Activities')
INSERT INTO skills (name, category_id, description) VALUES
    ('Hiking', (SELECT id FROM cat), 'Trail hiking and navigation'),
    ('Camping', (SELECT id FROM cat), 'Camping skills and setup'),
    ('Backpacking', (SELECT id FROM cat), 'Multi-day wilderness trips'),
    ('Rock Climbing', (SELECT id FROM cat), 'Indoor and outdoor climbing'),
    ('Mountain Climbing', (SELECT id FROM cat), 'Mountaineering skills'),
    ('Trail Running', (SELECT id FROM cat), 'Off-road running'),
    ('Mountain Biking', (SELECT id FROM cat), 'Off-road cycling'),
    ('Skiing', (SELECT id FROM cat), 'Downhill and cross-country skiing'),
    ('Snowboarding', (SELECT id FROM cat), 'Snowboard techniques'),
    ('Ice Skating', (SELECT id FROM cat), 'Ice skating skills'),
    ('Fishing', (SELECT id FROM cat), 'Freshwater and saltwater fishing'),
    ('Hunting', (SELECT id FROM cat), 'Hunting techniques and safety'),
    ('Orienteering', (SELECT id FROM cat), 'Navigation and map reading'),
    ('Survival Skills', (SELECT id FROM cat), 'Wilderness survival'),
    ('Birdwatching', (SELECT id FROM cat), 'Bird identification and watching'),
    ('Geocaching', (SELECT id FROM cat), 'GPS treasure hunting'),
    ('Skateboarding', (SELECT id FROM cat), 'Skateboard tricks and riding'),
    ('Rollerblading', (SELECT id FROM cat), 'Inline skating'),
    ('Parkour', (SELECT id FROM cat), 'Urban movement and freerunning')
ON CONFLICT (category_id, name) DO NOTHING;

-- Cooking Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Cooking')
INSERT INTO skills (name, category_id, description) VALUES
    ('Home Cooking', (SELECT id FROM cat), 'Everyday meal preparation'),
    ('Professional Cooking', (SELECT id FROM cat), 'Restaurant-level cooking'),
    ('Meal Prep', (SELECT id FROM cat), 'Batch cooking and planning'),
    ('Grilling & BBQ', (SELECT id FROM cat), 'Outdoor grilling techniques'),
    ('Sous Vide', (SELECT id FROM cat), 'Precision cooking method'),
    ('Italian Cuisine', (SELECT id FROM cat), 'Italian cooking traditions'),
    ('French Cuisine', (SELECT id FROM cat), 'French culinary techniques'),
    ('Asian Cuisine', (SELECT id FROM cat), 'Asian cooking styles'),
    ('Mexican Cuisine', (SELECT id FROM cat), 'Mexican food preparation'),
    ('Indian Cuisine', (SELECT id FROM cat), 'Indian cooking and spices'),
    ('Mediterranean Cuisine', (SELECT id FROM cat), 'Mediterranean diet cooking'),
    ('Vegetarian Cooking', (SELECT id FROM cat), 'Plant-based meal preparation'),
    ('Vegan Cooking', (SELECT id FROM cat), 'Fully plant-based cooking'),
    ('Sushi Making', (SELECT id FROM cat), 'Japanese sushi preparation'),
    ('Knife Skills', (SELECT id FROM cat), 'Professional cutting techniques'),
    ('Food Preservation', (SELECT id FROM cat), 'Canning, pickling, and fermenting')
ON CONFLICT (category_id, name) DO NOTHING;

-- Baking Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Baking')
INSERT INTO skills (name, category_id, description) VALUES
    ('Bread Baking', (SELECT id FROM cat), 'Artisan bread making'),
    ('Pastry Making', (SELECT id FROM cat), 'Creating pastries and croissants'),
    ('Cake Decorating', (SELECT id FROM cat), 'Decorating cakes and cupcakes'),
    ('Cookie Baking', (SELECT id FROM cat), 'Making various cookies'),
    ('Pie Making', (SELECT id FROM cat), 'Baking pies and tarts'),
    ('Chocolate Work', (SELECT id FROM cat), 'Working with chocolate'),
    ('Sourdough', (SELECT id FROM cat), 'Sourdough bread making'),
    ('Gluten-Free Baking', (SELECT id FROM cat), 'Baking without gluten'),
    ('Cake Making', (SELECT id FROM cat), 'Baking cakes from scratch')
ON CONFLICT (category_id, name) DO NOTHING;

-- Beverages Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Beverages')
INSERT INTO skills (name, category_id, description) VALUES
    ('Barista Skills', (SELECT id FROM cat), 'Espresso and coffee making'),
    ('Latte Art', (SELECT id FROM cat), 'Creating designs in coffee'),
    ('Bartending', (SELECT id FROM cat), 'Mixing drinks and cocktails'),
    ('Cocktail Making', (SELECT id FROM cat), 'Crafting cocktails'),
    ('Wine Knowledge', (SELECT id FROM cat), 'Wine tasting and pairing'),
    ('Beer Brewing', (SELECT id FROM cat), 'Home brewing beer'),
    ('Tea Preparation', (SELECT id FROM cat), 'Tea brewing and service'),
    ('Juice Making', (SELECT id FROM cat), 'Fresh juice preparation'),
    ('Smoothie Making', (SELECT id FROM cat), 'Blending healthy smoothies')
ON CONFLICT (category_id, name) DO NOTHING;

-- Nutrition Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Nutrition')
INSERT INTO skills (name, category_id, description) VALUES
    ('Nutrition Planning', (SELECT id FROM cat), 'Creating balanced meal plans'),
    ('Sports Nutrition', (SELECT id FROM cat), 'Nutrition for athletes'),
    ('Weight Management', (SELECT id FROM cat), 'Healthy weight strategies'),
    ('Diet Coaching', (SELECT id FROM cat), 'Coaching dietary changes'),
    ('Macros Tracking', (SELECT id FROM cat), 'Tracking nutritional macros'),
    ('Supplements Knowledge', (SELECT id FROM cat), 'Understanding supplements')
ON CONFLICT (category_id, name) DO NOTHING;

-- Visual Arts Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Visual Arts')
INSERT INTO skills (name, category_id, description) VALUES
    ('Drawing', (SELECT id FROM cat), 'Pencil and charcoal drawing'),
    ('Painting', (SELECT id FROM cat), 'Oil, acrylic, and watercolor'),
    ('Sculpture', (SELECT id FROM cat), 'Three-dimensional art'),
    ('Digital Art', (SELECT id FROM cat), 'Creating art digitally'),
    ('Portraiture', (SELECT id FROM cat), 'Portrait drawing and painting'),
    ('Landscape Art', (SELECT id FROM cat), 'Landscape painting'),
    ('Abstract Art', (SELECT id FROM cat), 'Non-representational art'),
    ('Street Art', (SELECT id FROM cat), 'Murals and graffiti art'),
    ('Calligraphy', (SELECT id FROM cat), 'Decorative handwriting'),
    ('Sketching', (SELECT id FROM cat), 'Quick drawing techniques'),
    ('Printmaking', (SELECT id FROM cat), 'Creating prints and editions')
ON CONFLICT (category_id, name) DO NOTHING;

-- Photography Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Photography')
INSERT INTO skills (name, category_id, description) VALUES
    ('Portrait Photography', (SELECT id FROM cat), 'Photographing people'),
    ('Landscape Photography', (SELECT id FROM cat), 'Nature and scenery'),
    ('Wedding Photography', (SELECT id FROM cat), 'Wedding event coverage'),
    ('Product Photography', (SELECT id FROM cat), 'Commercial product shots'),
    ('Street Photography', (SELECT id FROM cat), 'Candid urban photography'),
    ('Wildlife Photography', (SELECT id FROM cat), 'Animal photography'),
    ('Food Photography', (SELECT id FROM cat), 'Food styling and photography'),
    ('Night Photography', (SELECT id FROM cat), 'Low-light photography'),
    ('Drone Photography', (SELECT id FROM cat), 'Aerial photography'),
    ('Photo Editing', (SELECT id FROM cat), 'Post-processing images'),
    ('Lightroom', (SELECT id FROM cat), 'Adobe Lightroom editing'),
    ('Film Photography', (SELECT id FROM cat), 'Analog film photography')
ON CONFLICT (category_id, name) DO NOTHING;

-- Film & Video Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Film & Video')
INSERT INTO skills (name, category_id, description) VALUES
    ('Video Production', (SELECT id FROM cat), 'Creating video content'),
    ('Video Editing', (SELECT id FROM cat), 'Editing video footage'),
    ('Cinematography', (SELECT id FROM cat), 'Camera work and lighting'),
    ('Documentary Making', (SELECT id FROM cat), 'Creating documentaries'),
    ('Animation', (SELECT id FROM cat), '2D and 3D animation'),
    ('Screenwriting', (SELECT id FROM cat), 'Writing for film and TV'),
    ('Film Directing', (SELECT id FROM cat), 'Directing film productions'),
    ('Sound Design', (SELECT id FROM cat), 'Creating audio for film'),
    ('Color Grading', (SELECT id FROM cat), 'Video color correction'),
    ('YouTube Content', (SELECT id FROM cat), 'Creating YouTube videos'),
    ('TikTok Content', (SELECT id FROM cat), 'Short-form video creation'),
    ('Livestreaming', (SELECT id FROM cat), 'Live video broadcasting'),
    ('After Effects', (SELECT id FROM cat), 'Motion graphics and VFX'),
    ('Premiere Pro', (SELECT id FROM cat), 'Adobe Premiere editing'),
    ('Final Cut Pro', (SELECT id FROM cat), 'Apple video editing'),
    ('DaVinci Resolve', (SELECT id FROM cat), 'Resolve editing and color')
ON CONFLICT (category_id, name) DO NOTHING;

-- Writing Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Writing')
INSERT INTO skills (name, category_id, description) VALUES
    ('Creative Writing', (SELECT id FROM cat), 'Fiction and creative prose'),
    ('Technical Writing', (SELECT id FROM cat), 'Documentation and manuals'),
    ('Blog Writing', (SELECT id FROM cat), 'Writing blog content'),
    ('Journalism', (SELECT id FROM cat), 'News and feature writing'),
    ('Editing & Proofreading', (SELECT id FROM cat), 'Reviewing and correcting text'),
    ('Grant Writing', (SELECT id FROM cat), 'Writing funding proposals'),
    ('Resume Writing', (SELECT id FROM cat), 'Creating effective resumes'),
    ('Academic Writing', (SELECT id FROM cat), 'Scholarly papers and research'),
    ('Ghostwriting', (SELECT id FROM cat), 'Writing for others'),
    ('Poetry', (SELECT id FROM cat), 'Writing poetry'),
    ('Storytelling', (SELECT id FROM cat), 'Narrative craft'),
    ('Content Writing', (SELECT id FROM cat), 'Web and marketing content')
ON CONFLICT (category_id, name) DO NOTHING;

-- Crafts Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Crafts')
INSERT INTO skills (name, category_id, description) VALUES
    ('Knitting', (SELECT id FROM cat), 'Creating knitted items'),
    ('Crocheting', (SELECT id FROM cat), 'Crochet techniques'),
    ('Sewing', (SELECT id FROM cat), 'Hand and machine sewing'),
    ('Embroidery', (SELECT id FROM cat), 'Decorative stitching'),
    ('Quilting', (SELECT id FROM cat), 'Making quilts'),
    ('Pottery', (SELECT id FROM cat), 'Creating ceramic pieces'),
    ('Candle Making', (SELECT id FROM cat), 'Crafting candles'),
    ('Soap Making', (SELECT id FROM cat), 'Handmade soap crafting'),
    ('Leatherworking', (SELECT id FROM cat), 'Working with leather'),
    ('Paper Crafts', (SELECT id FROM cat), 'Origami and paper art'),
    ('Scrapbooking', (SELECT id FROM cat), 'Creating memory books'),
    ('Macramé', (SELECT id FROM cat), 'Knotting techniques'),
    ('Weaving', (SELECT id FROM cat), 'Textile weaving'),
    ('Beading', (SELECT id FROM cat), 'Creating beaded items'),
    ('Resin Art', (SELECT id FROM cat), 'Epoxy resin crafts'),
    ('Flower Arranging', (SELECT id FROM cat), 'Floral design')
ON CONFLICT (category_id, name) DO NOTHING;

-- Fashion Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Fashion')
INSERT INTO skills (name, category_id, description) VALUES
    ('Fashion Design', (SELECT id FROM cat), 'Designing clothing'),
    ('Pattern Making', (SELECT id FROM cat), 'Creating sewing patterns'),
    ('Tailoring', (SELECT id FROM cat), 'Altering and fitting clothes'),
    ('Fashion Styling', (SELECT id FROM cat), 'Styling outfits'),
    ('Costume Design', (SELECT id FROM cat), 'Designing costumes'),
    ('Personal Shopping', (SELECT id FROM cat), 'Shopping for clients'),
    ('Wardrobe Consulting', (SELECT id FROM cat), 'Wardrobe organization'),
    ('Shoe Design', (SELECT id FROM cat), 'Designing footwear'),
    ('Accessories Design', (SELECT id FROM cat), 'Creating fashion accessories')
ON CONFLICT (category_id, name) DO NOTHING;

-- Music Skills (adding more to existing category)
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Music')
INSERT INTO skills (name, category_id, description) VALUES
    ('Guitar', (SELECT id FROM cat), 'Acoustic and electric guitar'),
    ('Drums', (SELECT id FROM cat), 'Drum kit and percussion'),
    ('Violin', (SELECT id FROM cat), 'Violin playing'),
    ('Saxophone', (SELECT id FROM cat), 'Saxophone performance'),
    ('Trumpet', (SELECT id FROM cat), 'Trumpet playing'),
    ('Flute', (SELECT id FROM cat), 'Flute performance'),
    ('Cello', (SELECT id FROM cat), 'Cello playing'),
    ('Ukulele', (SELECT id FROM cat), 'Ukulele playing'),
    ('Bass Guitar', (SELECT id FROM cat), 'Bass guitar playing'),
    ('DJ Skills', (SELECT id FROM cat), 'DJing and mixing'),
    ('Music Production', (SELECT id FROM cat), 'Creating and producing music'),
    ('Songwriting', (SELECT id FROM cat), 'Writing songs'),
    ('Music Theory', (SELECT id FROM cat), 'Understanding music fundamentals'),
    ('Beatmaking', (SELECT id FROM cat), 'Creating beats and instrumentals'),
    ('Audio Engineering', (SELECT id FROM cat), 'Recording and mixing audio'),
    ('Harmonica', (SELECT id FROM cat), 'Harmonica playing'),
    ('Accordion', (SELECT id FROM cat), 'Accordion playing'),
    ('Keyboard', (SELECT id FROM cat), 'Electronic keyboard playing')
ON CONFLICT (category_id, name) DO NOTHING;

-- Dance Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Dance')
INSERT INTO skills (name, category_id, description) VALUES
    ('Ballet', (SELECT id FROM cat), 'Classical ballet'),
    ('Hip Hop Dance', (SELECT id FROM cat), 'Hip hop and street dance'),
    ('Salsa', (SELECT id FROM cat), 'Latin salsa dancing'),
    ('Bachata', (SELECT id FROM cat), 'Dominican bachata'),
    ('Tango', (SELECT id FROM cat), 'Argentine tango'),
    ('Contemporary Dance', (SELECT id FROM cat), 'Modern contemporary'),
    ('Jazz Dance', (SELECT id FROM cat), 'Jazz dance techniques'),
    ('Ballroom Dancing', (SELECT id FROM cat), 'Partner ballroom dances'),
    ('Breakdancing', (SELECT id FROM cat), 'Breaking and b-boying'),
    ('Tap Dance', (SELECT id FROM cat), 'Tap dancing'),
    ('Line Dancing', (SELECT id FROM cat), 'Country line dancing'),
    ('Swing Dance', (SELECT id FROM cat), 'Swing and lindy hop'),
    ('Belly Dancing', (SELECT id FROM cat), 'Middle Eastern dance'),
    ('Flamenco', (SELECT id FROM cat), 'Spanish flamenco'),
    ('K-Pop Dance', (SELECT id FROM cat), 'Korean pop choreography'),
    ('Choreography', (SELECT id FROM cat), 'Creating dance routines'),
    ('Zumba', (SELECT id FROM cat), 'Zumba fitness dance')
ON CONFLICT (category_id, name) DO NOTHING;

-- Theater Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Theater')
INSERT INTO skills (name, category_id, description) VALUES
    ('Acting', (SELECT id FROM cat), 'Performance and character work'),
    ('Voice Acting', (SELECT id FROM cat), 'Voice-over performance'),
    ('Improv', (SELECT id FROM cat), 'Improvisational theater'),
    ('Stand-up Comedy', (SELECT id FROM cat), 'Comedy performance'),
    ('Stage Direction', (SELECT id FROM cat), 'Directing theater productions'),
    ('Stage Design', (SELECT id FROM cat), 'Set and stage design'),
    ('Stage Lighting', (SELECT id FROM cat), 'Theater lighting design'),
    ('Costume Making', (SELECT id FROM cat), 'Creating theater costumes'),
    ('Stage Makeup', (SELECT id FROM cat), 'Theatrical makeup'),
    ('Puppetry', (SELECT id FROM cat), 'Puppet performance')
ON CONFLICT (category_id, name) DO NOTHING;

-- Voice Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Voice')
INSERT INTO skills (name, category_id, description) VALUES
    ('Singing', (SELECT id FROM cat), 'Vocal performance'),
    ('Vocal Training', (SELECT id FROM cat), 'Voice development'),
    ('Opera Singing', (SELECT id FROM cat), 'Classical opera'),
    ('Rap', (SELECT id FROM cat), 'Rapping and flow'),
    ('Beatboxing', (SELECT id FROM cat), 'Vocal percussion'),
    ('Choir Singing', (SELECT id FROM cat), 'Choral singing'),
    ('Vocal Harmony', (SELECT id FROM cat), 'Harmonizing vocals')
ON CONFLICT (category_id, name) DO NOTHING;

-- Teaching Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Teaching')
INSERT INTO skills (name, category_id, description) VALUES
    ('Tutoring', (SELECT id FROM cat), 'One-on-one instruction'),
    ('Curriculum Design', (SELECT id FROM cat), 'Creating learning plans'),
    ('Online Teaching', (SELECT id FROM cat), 'Virtual instruction'),
    ('Special Education', (SELECT id FROM cat), 'Teaching students with special needs'),
    ('ESL Teaching', (SELECT id FROM cat), 'Teaching English as a second language'),
    ('Math Tutoring', (SELECT id FROM cat), 'Teaching mathematics'),
    ('Science Tutoring', (SELECT id FROM cat), 'Teaching science subjects'),
    ('Test Preparation', (SELECT id FROM cat), 'SAT, GRE, and exam prep'),
    ('Music Teaching', (SELECT id FROM cat), 'Teaching music and instruments'),
    ('Sports Coaching', (SELECT id FROM cat), 'Coaching athletic skills')
ON CONFLICT (category_id, name) DO NOTHING;

-- Public Speaking Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Public Speaking')
INSERT INTO skills (name, category_id, description) VALUES
    ('Presentation Skills', (SELECT id FROM cat), 'Delivering presentations'),
    ('Keynote Speaking', (SELECT id FROM cat), 'Conference keynotes'),
    ('Debate', (SELECT id FROM cat), 'Debate and argumentation'),
    ('Podcast Hosting', (SELECT id FROM cat), 'Hosting podcasts'),
    ('MC & Hosting', (SELECT id FROM cat), 'Event hosting'),
    ('Toastmasters', (SELECT id FROM cat), 'Toastmasters techniques'),
    ('Speech Writing', (SELECT id FROM cat), 'Writing speeches')
ON CONFLICT (category_id, name) DO NOTHING;

-- Personal Development Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Personal Development')
INSERT INTO skills (name, category_id, description) VALUES
    ('Time Management', (SELECT id FROM cat), 'Managing time effectively'),
    ('Productivity', (SELECT id FROM cat), 'Getting things done'),
    ('Goal Setting', (SELECT id FROM cat), 'Setting and achieving goals'),
    ('Life Coaching', (SELECT id FROM cat), 'Personal coaching'),
    ('Career Coaching', (SELECT id FROM cat), 'Career development guidance'),
    ('Networking', (SELECT id FROM cat), 'Building professional networks'),
    ('Interview Skills', (SELECT id FROM cat), 'Job interview preparation'),
    ('Confidence Building', (SELECT id FROM cat), 'Building self-confidence'),
    ('Habit Formation', (SELECT id FROM cat), 'Creating positive habits'),
    ('Memory Techniques', (SELECT id FROM cat), 'Improving memory'),
    ('Speed Reading', (SELECT id FROM cat), 'Reading faster')
ON CONFLICT (category_id, name) DO NOTHING;

-- Parenting Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Parenting')
INSERT INTO skills (name, category_id, description) VALUES
    ('Childcare', (SELECT id FROM cat), 'Caring for children'),
    ('Newborn Care', (SELECT id FROM cat), 'Caring for infants'),
    ('Child Development', (SELECT id FROM cat), 'Understanding child growth'),
    ('Homeschooling', (SELECT id FROM cat), 'Teaching children at home'),
    ('Positive Discipline', (SELECT id FROM cat), 'Effective parenting techniques'),
    ('Baby Sign Language', (SELECT id FROM cat), 'Communicating with babies'),
    ('Kids Activities', (SELECT id FROM cat), 'Planning activities for children')
ON CONFLICT (category_id, name) DO NOTHING;

-- Health Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Health')
INSERT INTO skills (name, category_id, description) VALUES
    ('First Aid', (SELECT id FROM cat), 'Emergency first aid'),
    ('CPR', (SELECT id FROM cat), 'Cardiopulmonary resuscitation'),
    ('Massage Therapy', (SELECT id FROM cat), 'Therapeutic massage'),
    ('Physical Therapy', (SELECT id FROM cat), 'Rehabilitation exercises'),
    ('Elderly Care', (SELECT id FROM cat), 'Caring for seniors'),
    ('Medical Knowledge', (SELECT id FROM cat), 'General medical information'),
    ('Herbalism', (SELECT id FROM cat), 'Herbal remedies and medicine'),
    ('Aromatherapy', (SELECT id FROM cat), 'Essential oils and therapy')
ON CONFLICT (category_id, name) DO NOTHING;

-- Mental Wellness Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Mental Wellness')
INSERT INTO skills (name, category_id, description) VALUES
    ('Meditation', (SELECT id FROM cat), 'Mindfulness meditation'),
    ('Mindfulness', (SELECT id FROM cat), 'Present moment awareness'),
    ('Stress Management', (SELECT id FROM cat), 'Coping with stress'),
    ('Breathwork', (SELECT id FROM cat), 'Breathing techniques'),
    ('Journaling', (SELECT id FROM cat), 'Reflective writing practice'),
    ('Sleep Optimization', (SELECT id FROM cat), 'Improving sleep quality'),
    ('Anxiety Management', (SELECT id FROM cat), 'Managing anxiety'),
    ('Emotional Intelligence', (SELECT id FROM cat), 'Understanding emotions')
ON CONFLICT (category_id, name) DO NOTHING;

-- Beauty Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Beauty')
INSERT INTO skills (name, category_id, description) VALUES
    ('Makeup Artistry', (SELECT id FROM cat), 'Professional makeup application'),
    ('Hair Styling', (SELECT id FROM cat), 'Styling hair'),
    ('Hair Cutting', (SELECT id FROM cat), 'Cutting and trimming hair'),
    ('Hair Coloring', (SELECT id FROM cat), 'Dyeing and coloring hair'),
    ('Nail Art', (SELECT id FROM cat), 'Manicure and nail design'),
    ('Skincare', (SELECT id FROM cat), 'Skin care routines'),
    ('Eyebrow Styling', (SELECT id FROM cat), 'Shaping eyebrows'),
    ('Eyelash Extensions', (SELECT id FROM cat), 'Applying lash extensions'),
    ('Bridal Makeup', (SELECT id FROM cat), 'Wedding makeup'),
    ('Special Effects Makeup', (SELECT id FROM cat), 'SFX and theatrical makeup')
ON CONFLICT (category_id, name) DO NOTHING;

-- Gardening Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Gardening')
INSERT INTO skills (name, category_id, description) VALUES
    ('Vegetable Gardening', (SELECT id FROM cat), 'Growing vegetables'),
    ('Flower Gardening', (SELECT id FROM cat), 'Growing ornamental plants'),
    ('Indoor Plants', (SELECT id FROM cat), 'Houseplant care'),
    ('Landscaping', (SELECT id FROM cat), 'Designing outdoor spaces'),
    ('Lawn Care', (SELECT id FROM cat), 'Maintaining lawns'),
    ('Permaculture', (SELECT id FROM cat), 'Sustainable gardening'),
    ('Hydroponics', (SELECT id FROM cat), 'Soilless growing systems'),
    ('Composting', (SELECT id FROM cat), 'Creating compost'),
    ('Herb Gardening', (SELECT id FROM cat), 'Growing culinary herbs'),
    ('Fruit Tree Care', (SELECT id FROM cat), 'Maintaining fruit trees'),
    ('Bonsai', (SELECT id FROM cat), 'Miniature tree cultivation'),
    ('Greenhouse Growing', (SELECT id FROM cat), 'Greenhouse gardening')
ON CONFLICT (category_id, name) DO NOTHING;

-- Home Improvement Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Home Improvement')
INSERT INTO skills (name, category_id, description) VALUES
    ('Painting Walls', (SELECT id FROM cat), 'Interior and exterior painting'),
    ('Wallpaper Installation', (SELECT id FROM cat), 'Hanging wallpaper'),
    ('Window Installation', (SELECT id FROM cat), 'Installing windows'),
    ('Door Installation', (SELECT id FROM cat), 'Installing doors'),
    ('Fence Building', (SELECT id FROM cat), 'Building fences'),
    ('Shelf Installation', (SELECT id FROM cat), 'Installing shelving'),
    ('Basic Repairs', (SELECT id FROM cat), 'General home repairs'),
    ('HVAC Basics', (SELECT id FROM cat), 'Heating and cooling basics'),
    ('Appliance Repair', (SELECT id FROM cat), 'Fixing home appliances'),
    ('Caulking & Sealing', (SELECT id FROM cat), 'Waterproofing and sealing')
ON CONFLICT (category_id, name) DO NOTHING;

-- Interior Design Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Interior Design')
INSERT INTO skills (name, category_id, description) VALUES
    ('Space Planning', (SELECT id FROM cat), 'Optimizing room layouts'),
    ('Color Theory', (SELECT id FROM cat), 'Understanding color in design'),
    ('Furniture Selection', (SELECT id FROM cat), 'Choosing furniture'),
    ('Home Staging', (SELECT id FROM cat), 'Staging homes for sale'),
    ('Feng Shui', (SELECT id FROM cat), 'Chinese space arrangement'),
    ('Lighting Design', (SELECT id FROM cat), 'Planning interior lighting'),
    ('Textile Selection', (SELECT id FROM cat), 'Choosing fabrics and textiles'),
    ('Minimalist Design', (SELECT id FROM cat), 'Minimalist aesthetics')
ON CONFLICT (category_id, name) DO NOTHING;

-- Cleaning Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Cleaning')
INSERT INTO skills (name, category_id, description) VALUES
    ('Home Cleaning', (SELECT id FROM cat), 'Residential cleaning'),
    ('Deep Cleaning', (SELECT id FROM cat), 'Thorough cleaning'),
    ('Organization', (SELECT id FROM cat), 'Organizing spaces'),
    ('Decluttering', (SELECT id FROM cat), 'Reducing clutter'),
    ('Move-in/Move-out Cleaning', (SELECT id FROM cat), 'Transition cleaning'),
    ('Carpet Cleaning', (SELECT id FROM cat), 'Cleaning carpets'),
    ('Window Cleaning', (SELECT id FROM cat), 'Washing windows'),
    ('Pressure Washing', (SELECT id FROM cat), 'Power washing surfaces'),
    ('Green Cleaning', (SELECT id FROM cat), 'Eco-friendly cleaning')
ON CONFLICT (category_id, name) DO NOTHING;

-- Science Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Science')
INSERT INTO skills (name, category_id, description) VALUES
    ('Biology', (SELECT id FROM cat), 'Life sciences'),
    ('Chemistry', (SELECT id FROM cat), 'Chemical sciences'),
    ('Physics', (SELECT id FROM cat), 'Physical sciences'),
    ('Astronomy', (SELECT id FROM cat), 'Study of space'),
    ('Geology', (SELECT id FROM cat), 'Earth sciences'),
    ('Marine Biology', (SELECT id FROM cat), 'Ocean life sciences'),
    ('Botany', (SELECT id FROM cat), 'Plant sciences'),
    ('Research Methods', (SELECT id FROM cat), 'Scientific research'),
    ('Lab Techniques', (SELECT id FROM cat), 'Laboratory skills')
ON CONFLICT (category_id, name) DO NOTHING;

-- Engineering Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Engineering')
INSERT INTO skills (name, category_id, description) VALUES
    ('Mechanical Engineering', (SELECT id FROM cat), 'Mechanical systems'),
    ('Electrical Engineering', (SELECT id FROM cat), 'Electrical systems'),
    ('Civil Engineering', (SELECT id FROM cat), 'Infrastructure design'),
    ('Chemical Engineering', (SELECT id FROM cat), 'Chemical processes'),
    ('CAD Design', (SELECT id FROM cat), 'Computer-aided design'),
    ('AutoCAD', (SELECT id FROM cat), 'AutoCAD software'),
    ('SolidWorks', (SELECT id FROM cat), '3D CAD modeling'),
    ('Arduino', (SELECT id FROM cat), 'Arduino programming'),
    ('Raspberry Pi', (SELECT id FROM cat), 'Raspberry Pi projects'),
    ('Electronics', (SELECT id FROM cat), 'Electronic circuits')
ON CONFLICT (category_id, name) DO NOTHING;

-- Environment Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Environment')
INSERT INTO skills (name, category_id, description) VALUES
    ('Sustainability', (SELECT id FROM cat), 'Sustainable practices'),
    ('Recycling', (SELECT id FROM cat), 'Waste recycling'),
    ('Zero Waste Living', (SELECT id FROM cat), 'Minimizing waste'),
    ('Solar Energy', (SELECT id FROM cat), 'Solar power systems'),
    ('Renewable Energy', (SELECT id FROM cat), 'Clean energy sources'),
    ('Environmental Education', (SELECT id FROM cat), 'Teaching environmental topics'),
    ('Conservation', (SELECT id FROM cat), 'Nature conservation'),
    ('Urban Farming', (SELECT id FROM cat), 'City agriculture')
ON CONFLICT (category_id, name) DO NOTHING;

-- Gaming Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Gaming')
INSERT INTO skills (name, category_id, description) VALUES
    ('Esports', (SELECT id FROM cat), 'Competitive gaming'),
    ('FPS Games', (SELECT id FROM cat), 'First-person shooter games'),
    ('MOBA Games', (SELECT id FROM cat), 'Multiplayer online battle arena'),
    ('Strategy Games', (SELECT id FROM cat), 'Real-time and turn-based strategy'),
    ('Fighting Games', (SELECT id FROM cat), 'Fighting game techniques'),
    ('Speedrunning', (SELECT id FROM cat), 'Completing games quickly'),
    ('Game Streaming', (SELECT id FROM cat), 'Streaming gameplay'),
    ('Game Coaching', (SELECT id FROM cat), 'Coaching gamers'),
    ('Retro Gaming', (SELECT id FROM cat), 'Classic video games'),
    ('VR Gaming', (SELECT id FROM cat), 'Virtual reality gaming')
ON CONFLICT (category_id, name) DO NOTHING;

-- Board Games Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Board Games')
INSERT INTO skills (name, category_id, description) VALUES
    ('Chess', (SELECT id FROM cat), 'Chess strategy and tactics'),
    ('Poker', (SELECT id FROM cat), 'Poker strategy'),
    ('Bridge', (SELECT id FROM cat), 'Contract bridge'),
    ('Go', (SELECT id FROM cat), 'The game of Go'),
    ('Dungeons & Dragons', (SELECT id FROM cat), 'D&D and tabletop RPGs'),
    ('Board Game Design', (SELECT id FROM cat), 'Creating board games'),
    ('Magic: The Gathering', (SELECT id FROM cat), 'MTG card game'),
    ('Mahjong', (SELECT id FROM cat), 'Mahjong tile game')
ON CONFLICT (category_id, name) DO NOTHING;

-- Pet Care Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Pet Care')
INSERT INTO skills (name, category_id, description) VALUES
    ('Dog Training', (SELECT id FROM cat), 'Training dogs'),
    ('Cat Care', (SELECT id FROM cat), 'Caring for cats'),
    ('Pet Grooming', (SELECT id FROM cat), 'Grooming pets'),
    ('Aquarium Keeping', (SELECT id FROM cat), 'Maintaining fish tanks'),
    ('Bird Care', (SELECT id FROM cat), 'Caring for pet birds'),
    ('Reptile Care', (SELECT id FROM cat), 'Caring for reptiles'),
    ('Horse Care', (SELECT id FROM cat), 'Equine care'),
    ('Horseback Riding', (SELECT id FROM cat), 'Riding horses'),
    ('Pet Sitting', (SELECT id FROM cat), 'Caring for pets temporarily'),
    ('Pet First Aid', (SELECT id FROM cat), 'Emergency pet care')
ON CONFLICT (category_id, name) DO NOTHING;

-- Agriculture Skills
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Agriculture')
INSERT INTO skills (name, category_id, description) VALUES
    ('Farming', (SELECT id FROM cat), 'Agricultural farming'),
    ('Livestock Care', (SELECT id FROM cat), 'Caring for farm animals'),
    ('Beekeeping', (SELECT id FROM cat), 'Keeping bees'),
    ('Poultry Keeping', (SELECT id FROM cat), 'Raising chickens'),
    ('Dairy Farming', (SELECT id FROM cat), 'Dairy production'),
    ('Crop Management', (SELECT id FROM cat), 'Managing crops'),
    ('Irrigation', (SELECT id FROM cat), 'Water management for crops'),
    ('Tractors & Equipment', (SELECT id FROM cat), 'Farm machinery operation')
ON CONFLICT (category_id, name) DO NOTHING;

-- Languages Skills (adding more to existing category)
WITH cat AS (SELECT id FROM skill_categories WHERE name = 'Languages')
INSERT INTO skills (name, category_id, description) VALUES
    ('French', (SELECT id FROM cat), 'French language'),
    ('German', (SELECT id FROM cat), 'German language'),
    ('Italian', (SELECT id FROM cat), 'Italian language'),
    ('Portuguese', (SELECT id FROM cat), 'Portuguese language'),
    ('Mandarin Chinese', (SELECT id FROM cat), 'Mandarin Chinese language'),
    ('Japanese', (SELECT id FROM cat), 'Japanese language'),
    ('Korean', (SELECT id FROM cat), 'Korean language'),
    ('Arabic', (SELECT id FROM cat), 'Arabic language'),
    ('Russian', (SELECT id FROM cat), 'Russian language'),
    ('Hindi', (SELECT id FROM cat), 'Hindi language'),
    ('Dutch', (SELECT id FROM cat), 'Dutch language'),
    ('Swedish', (SELECT id FROM cat), 'Swedish language'),
    ('Polish', (SELECT id FROM cat), 'Polish language'),
    ('Turkish', (SELECT id FROM cat), 'Turkish language'),
    ('Greek', (SELECT id FROM cat), 'Greek language'),
    ('Hebrew', (SELECT id FROM cat), 'Hebrew language'),
    ('Vietnamese', (SELECT id FROM cat), 'Vietnamese language'),
    ('Thai', (SELECT id FROM cat), 'Thai language'),
    ('Sign Language', (SELECT id FROM cat), 'Sign language communication'),
    ('Latin', (SELECT id FROM cat), 'Classical Latin')
ON CONFLICT (category_id, name) DO NOTHING;

-- +goose Down
-- Note: This down migration removes only the new skills and categories added by this migration
-- It preserves the original Development, Languages, and Music categories and their skills

-- Remove skills from new categories
DELETE FROM skills WHERE category_id IN (
    SELECT id FROM skill_categories WHERE name IN (
        'Technology', 'Data & Analytics', 'Design', 'Business', 'Marketing', 'Finance',
        'Construction', 'Electrical', 'Plumbing', 'Automotive', 'Woodworking', 'Metalworking',
        'Team Sports', 'Individual Sports', 'Fitness', 'Martial Arts', 'Water Sports', 'Outdoor Activities',
        'Cooking', 'Baking', 'Beverages', 'Nutrition',
        'Visual Arts', 'Photography', 'Film & Video', 'Writing', 'Crafts', 'Fashion',
        'Dance', 'Theater', 'Voice',
        'Teaching', 'Public Speaking', 'Personal Development', 'Parenting',
        'Health', 'Mental Wellness', 'Beauty',
        'Gardening', 'Home Improvement', 'Interior Design', 'Cleaning',
        'Science', 'Engineering', 'Environment',
        'Gaming', 'Board Games',
        'Pet Care', 'Agriculture'
    )
);

-- Remove new skills added to existing Music category
DELETE FROM skills WHERE name IN (
    'Guitar', 'Drums', 'Violin', 'Saxophone', 'Trumpet', 'Flute', 'Cello', 'Ukulele',
    'Bass Guitar', 'DJ Skills', 'Music Production', 'Songwriting', 'Music Theory',
    'Beatmaking', 'Audio Engineering', 'Harmonica', 'Accordion', 'Keyboard'
) AND category_id = (SELECT id FROM skill_categories WHERE name = 'Music');

-- Remove new skills added to existing Languages category
DELETE FROM skills WHERE name IN (
    'French', 'German', 'Italian', 'Portuguese', 'Mandarin Chinese', 'Japanese',
    'Korean', 'Arabic', 'Russian', 'Hindi', 'Dutch', 'Swedish', 'Polish', 'Turkish',
    'Greek', 'Hebrew', 'Vietnamese', 'Thai', 'Sign Language', 'Latin'
) AND category_id = (SELECT id FROM skill_categories WHERE name = 'Languages');

-- Remove new categories
DELETE FROM skill_categories WHERE name IN (
    'Technology', 'Data & Analytics', 'Design', 'Business', 'Marketing', 'Finance',
    'Construction', 'Electrical', 'Plumbing', 'Automotive', 'Woodworking', 'Metalworking',
    'Team Sports', 'Individual Sports', 'Fitness', 'Martial Arts', 'Water Sports', 'Outdoor Activities',
    'Cooking', 'Baking', 'Beverages', 'Nutrition',
    'Visual Arts', 'Photography', 'Film & Video', 'Writing', 'Crafts', 'Fashion',
    'Dance', 'Theater', 'Voice',
    'Teaching', 'Public Speaking', 'Personal Development', 'Parenting',
    'Health', 'Mental Wellness', 'Beauty',
    'Gardening', 'Home Improvement', 'Interior Design', 'Cleaning',
    'Science', 'Engineering', 'Environment',
    'Gaming', 'Board Games',
    'Pet Care', 'Agriculture'
);
