import logo from '../logo.svg';

const Header = ({searchQuery, setSearchQuery, onChange}) => {
    return (
        <div className='row header'>
            <div className='col-9'>
                <img src={logo} className='logo' alt='logo' />
            </div>
            <div className='col-3'>
                <input
                    type='text'
                    className='search_bar'
                    placeholder='Search'
                    onChange={onChange}>
                </input>
            </div>
        </div>
    );
}

export default Header
