import mimetypes
import os

CREDS_FROM_FILE = os.environ.get('CREDS_FROM_FILE', False)
if CREDS_FROM_FILE:
    os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "kawalc1-google-credentials.json"
SECRET = os.environ.get('KAWALC1_SECRET', 'test')
AUTHENTICATION_ENABLED = SECRET is not 'test'
BASE_DIR = "."
TARGET_EXTENSION = ".webp"
PADDING_INNER = 2
PADDING_OUTER = 16

import logging

mpl_logger = logging.getLogger('matplotlib')
mpl_logger.setLevel(logging.WARNING)

LOCAL = True
INMEMORYSTORAGE_PERSIST = True
STATIC_DIR = os.path.join(BASE_DIR, 'static')
TRANSFORMED_DIR = os.environ.get('TRANSFORMED_DIR', STATIC_DIR + '/2024')
LOGS_PATH = os.environ.get('LOGS_PATH', BASE_DIR)
VALIDATION_DIR = os.path.join(BASE_DIR, 'validation')
DATASET_DIR = os.path.join(STATIC_DIR, 'datasets')
CONFIG_FILE = os.path.join(DATASET_DIR, 'gubernur-jakarta.json')
CATEGORIES_COUNT = 11
GS_BUCKET_NAME = 'kawalc1'
GS_FILE_OVERWRITE = True
GS_DEFAULT_ACL = 'publicRead'
VERSION = 'v0.9.1-alpha'