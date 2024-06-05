import logging
import os
import urllib
import pathlib

from io import BytesIO
import cv2
from os import path
import numpy as np
from urllib import parse
import urllib.request
import certifi
from PIL import Image
import settings

def is_url(file_path):
    return parse.urlparse(file_path).scheme in ('http', 'https',)


def _to_grayscale_image(input_stream):
    image = np.asarray(bytearray(input_stream.read()), dtype="uint8")
    input_stream.close()
    return cv2.imdecode(image, cv2.IMREAD_GRAYSCALE)


def _to_color_image(input_stream):
    pil_image = Image.open(BytesIO(input_stream.read()))
    input_stream.close()
    return pil_image.convert('RGB')


def read_color_image(file_path):
    file = read_file(file_path)
    return _to_color_image(file)


def _from_webp(input_stream):
    pil_image = Image.open(BytesIO(input_stream.read()))
    input_stream.close()
    return cv2.cvtColor(np.array(pil_image), cv2.COLOR_BGR2GRAY)


def read_image(file_path):
    filename, file_extension = os.path.splitext(file_path)
    file = read_file(file_path)
    if file_extension.lower() == ".webp":
        return _from_webp(file)
    return _to_grayscale_image(file)


def read_file(file_path):
    if is_url(file_path):
        logging.info("downloading from %s", file_path)
        return urllib.request.urlopen(file_path, cafile=certifi.where())
    else:
        logging.info("reading %s", os.path.abspath(file_path))
        try:
            return open(file_path, 'rb')
        except Exception as e:
            logging.error("Could not open %s \n %s", os.path.abspath(file_path), e)
            raise e


def write_string(file_path, file_name, string):
    pathlib.Path(file_path).mkdir(parents=True, exist_ok=True)
    full_path = os.path.join(file_path, file_name)
    print(full_path)
    with open(full_path, "w") as text_file:
        text_file.write(string)


def write_image(file_path, image):
    file, extension = path.splitext(file_path)
    if extension.lower() == ".webp":
        logging.info("writing .webp %s", file_path)
        im_pil = Image.fromarray(image)
        im_pil.save(file_path, 'webp')
    else:
        logging.info("writing image %s %s", file_path, extension)
        cv2.imwrite(file_path, image)


def write_color_image(file_path, image):
    logging.info("writing color image %s", file_path)
    image.save(file_path, 'webp')


def write_json(file_path, json_data):
    logging.info(f"writing json {file_path}")
    with open(file_path, 'w') as json_file:
        json_file.write(json_data)


def image_url(file_path):
    return file_path.replace("static/transformed", "../static/transformed") if settings.LOCAL else file_path
