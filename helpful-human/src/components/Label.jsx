const Label = ({label}) => {
    return (
        <label
            id={'label' + label.text}
            htmlFor={label.text}
            style={label.style}
        >
            {label.text}
        </label>
    );
}

export default Label
