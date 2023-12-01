import storage from 'local-storage'
import './App.css';
import {useState} from 'react'
import ReactPaginate from 'react-paginate'
import Header  from './components/Header'
import Content from './components/Content'
import Button  from './components/Button'

// TODO: bundle different funcitonality into separate respective functions/vars
function App() {
    let allColors = generateColors();
    const [colors, setColors] = useState(allColors);
    const [searchQuery, setSearchQuery] = useState('');

    let currPageNum = 0;
    let pageSize = 12;
    let numPages = Math.ceil(colors.length / pageSize);

    const [swatches, setSwatches] = useState(colors.slice(0,pageSize).map((item) => ({
        backgroundColor: item.backgroundColor,
        width: '220px'
    })));

    const [labels, setLabels] = useState(swatches.map((item) => ({
        text: item.backgroundColor,
        style: {
            fontSize: '1.5rem',
            lineHeight: '50px',
            marginTop: '207px',
            padding: '0rem 1.5rem'
        }
    })));

    function onChangeSearch(e) {
        setSearchQuery(e.target.value);
        if (searchQuery !== '') {
            setColors(allColors.filter(color => color.backgroundColor.includes(searchQuery)));
        } else {
            setColors(allColors);
        }

        setSwatches(colors.slice(0,pageSize).map((item) => ({
            backgroundColor: item.backgroundColor,
            width: '220px'
        })));

        setLabels(colors.map((item) => ({
            text: item.backgroundColor,
            style: {
                fontSize: '1.5rem',
                lineHeight: '50px',
                marginTop: '207px',
                padding: '0rem 1.5rem'
            }
        })));
    }

    function selectPage({selected}) {
        currPageNum = selected;
        let newColors = colors.slice(pageSize * currPageNum,
                                     pageSize * (currPageNum + 1));
        let newSwatches = [];
        newColors.forEach((color) => newSwatches.push({
            backgroundColor: color.backgroundColor,
            width: '220px'
        }));

        let newLabels = [];
        newColors.forEach((color) => newLabels.push({
            text: color.backgroundColor,
            style: {
                fontSize: '1.5rem',
                lineHeight: '50px',
                marginTop: '207px',
                padding: '0rem 1.5rem'
            }
        }));

        setSwatches(newSwatches);
        setLabels(newLabels);
    }

    const pagination = (
        <ReactPaginate
            previousLabel={'Previous'}
            nextLabel={'Next'}
            pageCount={numPages}
            onPageChange={selectPage}
            containerClassName={'pagination'}
            previousLinkClassName={'paginationPrevious'}
            nextLinkClassName={'paginationNext'}
            activeClassName={'paginationActive'}
            disabledClassName={'paginationDisabled'}
            pageRangeDisplayed={numPages}
        />
    );

    const [navRow, setNavRow] = useState(pagination);

    function onClickDetail(e) {
        let newSwatches = [];
        newSwatches.push({
            backgroundColor: e.target.id,
            width: '100%',
            height: '650px',
            borderColor: 'black'
        });

        let restSwatches = swatches.filter(item => item.backgroundColor !== e.target.id);
        restSwatches.forEach(item => newSwatches.push({
            backgroundColor: item.backgroundColor,
            width: '176px',
            height: '176px'
        }));

        let newLabels = [];
        let detailLabel = {
            text: e.target.id,
            style: {
                fontSize: '4rem',
                lineHeight: '150px',
                height: '150px',
                marginTop: '498px',
                padding: '0rem 4.5rem'
            }
        };
        newLabels.push(detailLabel);

        let restLabels = labels.filter(item => item.text !== e.target.id);
        restLabels.forEach(label => newLabels.push({
            text: label.text,
            style: {
                lineHeight: '65px',
                height: '65px',
                marginTop: '110px',
                padding: '0rem 1.0rem'
            }
        }));

        setSwatches(newSwatches.slice(0,6));
        setLabels(newLabels.slice(0,6));
        setNavRow(clearButton);
    }

    function onClickClear(e) {
        let newSwatches = [];
        let page = colors.slice(pageSize * currPageNum,
                                pageSize * (currPageNum + 1));
        newSwatches = page.map((item) => ({
            backgroundColor: item.backgroundColor,
            width: '220px',
            height: '258px'
        }));

        let newLabels = [];
        newLabels = page.map((item) => ({
            text: item.backgroundColor,
            style: {
                fontSize: '1.5rem',
                lineHeight: '50px',
                height: '50px',
                width: '100%',
                marginTop: '207px'
            }
        }));

        setSwatches(newSwatches);
        setLabels(newLabels);
        setNavRow(pagination);
    }

    function onClickRandom(e) {
        let x = Math.floor(Math.random() * swatches.length);
        let randomSwatch = {
            target: {
                id: swatches[x].backgroundColor
            }
        };
        onClickDetail(randomSwatch);
    }

    function onClickGroup(e) {
        // TODO: to be implemented
    }

    const clearButton = (
        <Button
            key='btnClear'
            className='clear_btn'
            valueName='Clear'
            onClick={onClickClear}
        />
    );

    return (
        <div className='container-fluid h-100'>
            <Header
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                onChange={onChangeSearch}
            />
            <Content
                swatches={swatches}
                labels={labels}
                navRow={navRow}
                onClickRandom={onClickRandom}
                onClickGroup={onClickGroup}
                onClickDetail={onClickDetail}
            />
        </div>
    );
}

function generateColors() {
    if (storage.get('colorList') === null) {
        let colors = [];
        for (let i = 0; i < 120; i++) {
            let color = '#'+(Math.random() * 0xFFFFFF << 0).toString(16).padStart(6, '0');
            colors.push({backgroundColor: color});
        }
        storage.set('colorList', colors);
    } else {
        return storage.get('colorList');
    }
}

export default App;
