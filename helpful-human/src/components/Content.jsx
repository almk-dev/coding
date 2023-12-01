import Sidebar from './Sidebar'
import View  from './View'

const Content = ({swatches, labels, navRow, onClickRandom, onClickGroup, onClickDetail}) => {
    return (
        <div className='row content'>
            <Sidebar
                onClickRandom={onClickRandom}
                onClickGroup={onClickGroup}
            />
            <View
                swatches={swatches}
                labels={labels}
                navRow={navRow}
                onClick={onClickDetail}
            />
        </div>
    );
}

export default Content
