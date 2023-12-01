import Label from './Label'

const Swatch = ({swatch, label, onClick}) => {
    return (
        <div
            id={swatch.backgroundColor}
            className='swatch'
            style={swatch}
            onClick={onClick}
        >
            <Label key={'label' + label.text} label={label} />
        </div>
    );
}

export default Swatch
