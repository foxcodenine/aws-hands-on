function calculateArea(length, width) {
    return length * width;
}

export const handler = async (event, context) => {
    // Get the length and width parameters from the event object. The
    // runtime converts the event JSON to a JavaScript object.
    const { length, width } = event;

    if (length === undefined || width === undefined) {
        throw new Error("Event must contain 'length' and 'width' keys");
    }

    const area = calculateArea(length, width);

    console.log(`The area is ${area}`);
    console.log('CloudWatch log group:', context.logGroupName);

    return JSON.stringify({ area });
};
