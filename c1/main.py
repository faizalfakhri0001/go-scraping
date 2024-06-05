import json
import logging
from os import path
import traceback

from mengenali import extraction, settings

def load_config(config_file_name):
    with open(path.join(settings.DATASET_DIR, config_file_name)) as config_file:
        config = json.load(config_file)
    return config

def extract_upload(filename, config_file):
    output = extraction.extract_rois(
        filename, 
        settings.STATIC_DIR, 
        path.join(settings.STATIC_DIR, 'extracted'),
        settings.STATIC_DIR,
        load_config(config_file)
    )
    return output

def main():
    import argparse

    parser = argparse.ArgumentParser(description='Extract and upload image data.')
    parser.add_argument('filename', type=str, help='Name of the file to process.')
    parser.add_argument('config_file', type=str, help='Configuration file to use.')

    args = parser.parse_args()

    try:
        result = extract_upload(args.filename, args.config_file)
        print(json.dumps(result, indent=4))
    except Exception as e:
        logging.error("An error occurred: %s", e)
        logging.error(traceback.format_exc())

if __name__ == "__main__":
    main()
