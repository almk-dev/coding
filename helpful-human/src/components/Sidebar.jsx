import Button  from './Button'

const Sidebar = ({onClickRandom, onClickGroup}) => {
    let groupList = ['Red','Orange','Yellow','Green','Blue','Purple','Brown','Gray'];

    return (
        <div className='col-3 sidebar'>
            <Button
                key='btnRandom'
                className='random_btn'
                valueName='Random Color'
                onClick={onClickRandom}
            />
            {groupList.map((groupName) => (
            <Button
                key={'group' + groupName}
                className='group_btn'
                valueName={groupName}
                onClick={onClickGroup}
            />
            ))}
        </div>
    );
}

export default Sidebar
