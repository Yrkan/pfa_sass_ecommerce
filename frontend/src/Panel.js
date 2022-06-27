import * as React from "react";
import { Admin, Resource, ListGuesser } from 'react-admin';
import jsonServerProvider from 'ra-data-json-server';


const Panel = ({link}) => {
    const dataProvider = jsonServerProvider(link);

    return <Admin dataProvider={dataProvider} >
        <Resource name="users" list={ListGuesser} />
    </Admin>
};

export default Panel;