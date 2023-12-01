import Swatch from './Swatch'

const View = ({swatches, labels, navRow, onClick}) => {
    return (
        <div className='col-9 view'>
            <div className='col-12 page'>
                {swatches.map((swatch) => (
                <Swatch
                    key={'swatch' + swatch.backgroundColor + swatch.width}
                    swatch={swatch}
                    label={labels.filter(item => item.text === swatch.backgroundColor)[0]}
                    onClick={onClick}
                />
                ))}
            </div>
            <div className='col-12 navrow'>
                {navRow}
            </div>
        </div>
    );
}

export default View
