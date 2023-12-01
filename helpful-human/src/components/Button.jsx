const Button = ({className, valueName, onClick}) => {
    return (
        <input
            type='button'
            className={className}
            value={valueName}
            onClick={onClick}
        />
    );
}

export default Button
