#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <FTR_CLIENT_PATH>"
    exit 1
fi

FTR_CLIENT_PATH=$1

echo "Seeding categories..."

cat ./scripts/categories.txt | ./scripts/create_categories.py http://localhost:8000 > /dev/null

echo "Seeding assets..."

echo -e "1\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic" > /dev/null
echo -e "2\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic" > /dev/null
echo -e "3\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic" > /dev/null

echo -e "4\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/BodyParts/Hair/Basic" > /dev/null
echo -e "5\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/BodyParts/Beard/Basic" > /dev/null
echo -e "6\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/BodyParts/Eyebrows/Basic" > /dev/null
echo -e "7\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/BodyParts/Eyes/Basic" > /dev/null
echo -e "8\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/BodyParts/Mouth/Basic" > /dev/null

echo -e "9\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/Equipment/Earrings/Common" > /dev/null
echo -e "10\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/Equipment/Back/Common" > /dev/null
echo -e "11\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/HeroEditor4D/Common/Sprites/Equipment/Mask/Common" > /dev/null

echo "Done."
