import json
import logging

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(event, context):
    if 'length' not in event or 'width' not in event:
        raise ValueError("Event must contain 'length' and 'width' keys")

    length = event['length']
    width = event['width']

    area = calculate_area(length, width)
    logger.info(f"The area is {area}")
    logger.info(f"CloudWatch logs group: {context.log_group_name}")

    return json.dumps({"area": area})

def calculate_area(length, width):
    return length * width
